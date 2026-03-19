package users

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

// isSequential checks for simple increasing or decreasing sequences
func isSequential(s string) bool {
	if len(s) < 5 {
		return false
	}
	runes := []rune(s)
	inc := true
	dec := true
	for i := 1; i < len(runes); i++ {
		if runes[i] != runes[i-1]+1 {
			inc = false
		}
		if runes[i] != runes[i-1]-1 {
			dec = false
		}
	}
	return inc || dec
}

func IsWeakPassword(password string) error {
	// check is password weak
	weakPasswords := map[string]struct{}{
		"123456":     {},
		"password":   {},
		"qwerty":     {},
		"111111":     {},
		"12345678":   {},
		"abc123":     {},
		"123456789":  {},
		"123123":     {},
		"qwertyuiop": {},
		"letmein":    {},
		"admin":      {},
		"welcome":    {},
		"monkey":     {},
		"passw0rd":   {},
		"starwars":   {},
	}

	if _, found := weakPasswords[password]; found {
		return errors.New("password is too weak (common password)")
	}

	// only digits
	onlyDigits := true
	for _, r := range password {
		if r < '0' || r > '9' {
			onlyDigits = false
			break
		}
	}
	if onlyDigits {
		return errors.New("password is too weak (only digits)")
	}

	// only letters
	onlyLetters := true
	for _, r := range password {
		if r < 'A' || (r > 'Z' && r < 'a') || r > 'z' {
			onlyLetters = false
			break
		}
	}
	if onlyLetters {
		return errors.New("password is too weak (only letters)")
	}

	// simple sequential patterns (e.g. abcdef, 123456)
	if isSequential(password) {
		return errors.New("password is too weak (sequential pattern)")
	}

	return nil
}

func checkPassword(password, passwordConfirm string) error {
	if subtle.ConstantTimeCompare([]byte(password), []byte(passwordConfirm)) != 1 {
		return fmt.Errorf("passwords do not match")
	}

	passLength := utf8.RuneCountInString(password)
	if passLength < 6 {
		return errors.New("password length must be greater or equal 6 symbols")
	}

	if passLength > 50 {
		return errors.New("password length must be less or equal 50 symbols")
	}

	if err := IsWeakPassword(password); err != nil {
		return err
	}

	return nil
}

func generateHashFromPassword(password string, cost int) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password is empty")
	}

	if cost == 0 {
		cost = bcrypt.DefaultCost
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("cannot hash password: %w", err)
	}

	return string(hashedPassword), nil
}
