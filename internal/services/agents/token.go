package agents

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
)

func ParseAuthHeader(h string) (agentID, secret string, err error) {
	const pfx = "Sentinel "
	if !strings.HasPrefix(h, pfx) {
		return "", "", errors.New("no Sentinel auth header")
	}
	tok := strings.TrimSpace(strings.TrimPrefix(h, pfx))
	if !strings.HasPrefix(tok, "snt_") {
		return "", "", errors.New("bad token prefix")
	}
	body := strings.TrimPrefix(tok, "snt_")
	dot := strings.IndexByte(body, '.')
	if dot <= 0 || dot == len(body)-1 {
		return "", "", errors.New("bad token format")
	}
	return body[:dot], body[dot+1:], nil
}

func NewAgentToken(agentID string) (token string, secret string, tokenHint string, err error) {
	if agentID == "" {
		return "", "", "", errors.New("agentID is empty")
	}
	sec := make([]byte, 32)
	if _, err = io.ReadFull(cryptorand.Reader, sec); err != nil {
		return "", "", "", fmt.Errorf("rand: %w", err)
	}
	secret = base64.RawURLEncoding.EncodeToString(sec)
	token = "snt_" + agentID + "." + secret

	if len(secret) >= 6 {
		tokenHint = "snt_" + agentID[:3] + "..." + secret[len(secret)-6:]
	} else {
		tokenHint = "snt_" + agentID[:3] + "..." + secret
	}
	return token, secret, tokenHint, nil
}

type Argon2Params struct {
	Time        uint32 // iterations
	MemoryKiB   uint32 // memory in KiB (64*1024 = 64 MiB)
	Parallelism uint8
	SaltLen     uint32
	KeyLen      uint32
}

// Default parameters: ~64 MiB, 3 iterations, 1 thread.
var DefaultArgon2 = Argon2Params{
	Time:        3,
	MemoryKiB:   64 * 1024,
	Parallelism: 1,
	SaltLen:     16,
	KeyLen:      32,
}

func HashSecretArgon2id(secret string, p Argon2Params) (string, error) {
	if secret == "" {
		return "", errors.New("empty secret")
	}
	salt := make([]byte, p.SaltLen)
	if _, err := io.ReadFull(cryptorand.Reader, salt); err != nil {
		return "", fmt.Errorf("salt rand: %w", err)
	}
	sum := argon2.IDKey([]byte(secret), salt, p.Time, p.MemoryKiB, p.Parallelism, p.KeyLen)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		p.MemoryKiB, p.Time, p.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(sum),
	), nil
}

// VerifySecret compares the secret with the stored hash
func VerifySecret(secret, stored string) (bool, error) {
	// expects: $argon2id$v=19$m=...,t=...,p=...$<salt>$<sum>
	if !strings.HasPrefix(stored, "$argon2id$") {
		return false, errors.New("bad argon2id header")
	}
	parts := strings.Split(stored, "$")
	// ["", "argon2id", "v=19", "m=..,t=..,p=..", "<salt>", "<sum>"]
	if len(parts) != 6 {
		return false, errors.New("bad argon2id format")
	}
	if parts[2] != "v=19" {
		return false, errors.New("unsupported argon2 version")
	}
	// m=..,t=..,p=..
	var mem uint32
	var tim uint32
	var par uint8

	for kv := range strings.SplitSeq(parts[3], ",") {
		kvp := strings.SplitN(kv, "=", 2)
		if len(kvp) != 2 {
			return false, errors.New("bad argon2 params")
		}
		switch kvp[0] {
		case "m":
			var v uint64
			_, err := fmt.Sscan(kvp[1], &v)
			if err != nil {
				return false, err
			}
			mem = uint32(v)
		case "t":
			var v uint64
			_, err := fmt.Sscan(kvp[1], &v)
			if err != nil {
				return false, err
			}
			tim = uint32(v)
		case "p":
			var v uint64
			_, err := fmt.Sscan(kvp[1], &v)
			if err != nil {
				return false, err
			}
			par = uint8(v)
		default:
			return false, errors.New("unknown argon2 param")
		}
	}

	saltB, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("salt decode: %w", err)
	}

	sumB, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("sum decode: %w", err)
	}

	got := argon2.IDKey([]byte(secret), saltB, tim, mem, par, uint32(len(sumB)))
	if len(got) != len(sumB) {
		return false, nil
	}

	// Checksum comparison without leaking timing info
	var v byte
	for i := range got {
		v |= got[i] ^ sumB[i]
	}

	return v == 0, nil
}
