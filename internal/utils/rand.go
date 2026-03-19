package utils

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const defaultAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-@#$%"

// GenerateRandomString generates a random string of the specified length using the provided alphabet.
func GenerateRandomString(n int, alphabet string) (string, error) {
	if n <= 0 {
		return "", nil
	}
	if alphabet == "" {
		alphabet = defaultAlphabet
	}
	// rejection sampling: берём байты до ближайшего кратного len(alphabet)
	alLen := byte(len(alphabet))
	maxrb := byte(255 - (256 % int(alLen)))

	out := make([]byte, n)
	i := 0
	buf := make([]byte, 64) // читаем пачками для эффективности

	for i < n {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		for _, rb := range buf {
			if rb > maxrb {
				continue // отбросить, чтобы не было смещения
			}
			out[i] = alphabet[int(rb)%int(alLen)]
			i++
			if i == n {
				break
			}
		}
	}

	return string(out), nil
}

// GenerateRandomNumber generates a random number of the specified length.
func GenerateRandomNumber(length uint) (int64, error) {
	if length == 0 {
		return 0, errors.New("length must be >= 1")
	}
	// int64 maximum 19 digits (9223372036854775807).
	if length > 19 {
		return 0, errors.New("length too large for int64; use string variant")
	}

	// min = 10^(length-1), max = 10^length - 1
	pow10 := func(n uint) int64 {
		var x int64 = 1
		for i := uint(0); i < n; i++ {
			x *= 10
		}
		return x
	}

	var min int64 = 1
	if length > 1 {
		min = pow10(length - 1)
	}
	max := pow10(length) - 1

	// range [min, max] inclusive
	rng := big.NewInt(max - min + 1)
	n, err := rand.Int(rand.Reader, rng)
	if err != nil {
		return 0, err
	}
	return n.Int64() + min, nil
}
