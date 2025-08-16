package certchecker

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	ErrNoPeerCertificate = errors.New("peer did not present certificate for domain")
	ErrExpired           = errors.New("certificate expired")
	ErrNotYetValid       = errors.New("certificate not yet valid")
	ErrInvalidHostname   = errors.New("invalid hostname")
	ErrNoCertificate     = errors.New("certificate serial not found")
)

var checkers struct {
	sync.Mutex
	state map[string]bool
}

func init() {
	checkers.state = make(map[string]bool)
}

type Domain struct {
	Domain string `json:"domain"`
	Port   string `json:"port"`
	cert   *x509.Certificate
}

type Status struct {
	Valid    bool   `json:"valid"`
	Err      string `json:"last_error"`
	Time     string `json:"last_check"`
	Validity int    `json:"valid_days"`
}

type Subject struct {
	CommonName         string   `json:"cn"`
	Country            []string `json:"c,omitempty"`
	Organization       []string `json:"o,omitempty"`
	OrganizationalUnit []string `json:"ou,omitempty"`
}

type Signature struct {
	Algorithm int    `json:"algorithm"`
	Value     string `json:"value"`
}

type Certificate struct {
	NotBefore      time.Time         `json:"not_before"`
	NotAfter       time.Time         `json:"not_after"`
	Issuer         Subject           `json:"issuer"`
	Subject        Subject           `json:"subject"`
	SerialNumber   string            `json:"serial"`
	AlternateNames []string          `json:"alternate_names,omitempty"`
	Signature      Signature         `json:"signature"`
	Fingerprints   map[string]string `json:"fingerprints"`
}

func toHexString(data []byte) string {
	result := make([]string, len(data))
	for i := 0; i < len(data); i += 1 {
		result[i] = hex.EncodeToString(data[i : i+1])
	}
	return strings.Join(result, ":")
}

func convertCert(cert *x509.Certificate) *Certificate {
	result := &Certificate{
		NotBefore:      cert.NotBefore.UTC(),
		NotAfter:       cert.NotAfter.UTC(),
		SerialNumber:   toHexString(cert.SerialNumber.Bytes()),
		AlternateNames: cert.DNSNames,
		Fingerprints:   make(map[string]string),
	}

	result.Signature = Signature{
		Value:     toHexString(cert.Signature),
		Algorithm: int(cert.SignatureAlgorithm),
	}

	result.Subject = Subject{
		CommonName:         cert.Subject.CommonName,
		Country:            cert.Subject.Country,
		Organization:       cert.Subject.Organization,
		OrganizationalUnit: cert.Subject.OrganizationalUnit,
	}

	result.Issuer = Subject{
		CommonName:         cert.Issuer.CommonName,
		Country:            cert.Issuer.Country,
		Organization:       cert.Issuer.Organization,
		OrganizationalUnit: cert.Issuer.OrganizationalUnit,
	}

	s256 := sha256.New()
	s256.Write(cert.Raw)
	result.Fingerprints["sha256"] = toHexString(s256.Sum(nil))

	s1 := sha1.New()
	s1.Write(cert.Raw)
	result.Fingerprints["sha1"] = toHexString(s1.Sum(nil))

	return result
}

func (d *Domain) GetCertificate() (*x509.Certificate, error) {
	// check if d.Domain is an IP address
	ip := net.ParseIP(d.Domain)
	if ip != nil { // d.Domain is an IP address
		return d.GetCertificateIP()
	} else { // d.Domain is a domain name
		return d.GetCertificateDomain()
	}
}

func (d *Domain) GetCertificateDomain() (*x509.Certificate, error) {
	// dial the remote server with timeout
	c, err := net.DialTimeout("tcp", d.Domain+":"+d.Port, time.Second*10)
	if err != nil {
		return nil, err
	}

	conn := tls.Client(c, &tls.Config{
		InsecureSkipVerify: true,     // we check expiration and hostname afterwars, we're only interested in the presented certificate
		ServerName:         d.Domain, // Set the ServerName to support checking vHost certs using SNI
	})
	if conn == nil {
		return nil, err
	}

	// make sure the handshake will timeout so the check will return
	// at some point
	if err := conn.SetDeadline(time.Now().Add(time.Second * 10)); err != nil {
		return nil, err
	}

	if err := conn.Handshake(); err != nil {
		return nil, err
	}

	state := conn.ConnectionState()
	for _, cert := range state.PeerCertificates {
		if ok := cert.VerifyHostname(d.Domain); ok == nil {
			return cert, nil
		}
	}

	return nil, ErrNoPeerCertificate
}

func (d *Domain) GetCertificateIP() (*x509.Certificate, error) {
	// dial the remote server with timeout
	c, err := net.DialTimeout("tcp", d.Domain+":"+d.Port, time.Second*10)
	if err != nil {
		return nil, err
	}

	conn := tls.Client(c, &tls.Config{
		InsecureSkipVerify: true, // skip hostname and trusted CA verification
		ServerName:         d.Domain,
	})
	if conn == nil {
		return nil, err
	}

	if err := conn.Handshake(); err != nil {
		return nil, ErrNoPeerCertificate // no certificate presented by peer
	}

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return nil, ErrNoPeerCertificate // no certificate presented by peer
	}

	return state.PeerCertificates[0], nil // return the first certificate in the list
}

func (d *Domain) Check() error {
	cert, err := d.GetCertificate()
	if err != nil {
		return err
	}
	d.cert = cert

	now := time.Now().UTC()
	if !now.Before(cert.NotAfter) {
		return ErrExpired
	}

	if !now.After(cert.NotBefore) {
		return ErrNotYetValid
	}

	return nil
}

func CheckDomain(domain, port string) {
	ticker := time.NewTicker(time.Minute * 5)
	log.Printf("starting domain checker for \"%s:%s\"\n", domain, port)

	checkers.Lock()
	checkers.state[domain+":"+port] = true
	checkers.Unlock()

	for {
		checkers.Lock()
		v, ok := checkers.state[domain+":"+port]
		checkers.Unlock()

		if !v || !ok {
			log.Printf("stopping check on \"%s:%s\"\n", domain, port)
			break
		}

		d := &Domain{
			Domain: domain,
			Port:   port,
		}

		start := time.Now()
		status := &Status{Time: start.UTC().Format(time.RFC3339)}
		err := d.Check()
		if err != nil {
			log.Printf("checking domain \"%s:%s\": %s\n", domain, port, err.Error())
			status.Valid = false
			status.Err = err.Error()
		} else {
			now := time.Now().UTC().Unix()
			validity := d.cert.NotAfter.Unix() - now
			status.Valid = true
			status.Validity = int(validity / 86400)
			log.Printf("checking domain \"%s:%s\": certificate is valid for %d days", domain, port, status.Validity)
		}

		// store latest check and certificate
		// d.Store(status)

		// wait for 5 minutes
		<-ticker.C
	}
}
