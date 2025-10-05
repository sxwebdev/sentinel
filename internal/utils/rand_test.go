package utils

import (
	"fmt"
	"strings"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		alphabet string
	}{
		{"length 0", 0, ""},
		{"length 1 default alphabet", 1, ""},
		{"length 10 default alphabet", 10, ""},
		{"length 100 default alphabet", 100, ""},
		{"length 1000 default alphabet", 1000, ""},
		{"length 100 custom alphabet", 100, "abc"},
		{"length 100 custom alphabet", 100, "0123456789"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := GenerateRandomString(tt.n, tt.alphabet)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			fmt.Println(s)
			if len(s) != tt.n {
				t.Errorf("expected length %d, got %d", tt.n, len(s))
			}
			if tt.alphabet != "" {
				for _, c := range s {
					if !strings.ContainsRune(tt.alphabet, c) {
						t.Errorf("character %q not in alphabet %q", c, tt.alphabet)
					}
				}
			}
		})
	}
}

func TestGenerateRandomNumber(t *testing.T) {
	tests := []struct {
		name   string
		length uint
	}{
		{"length 1", 1},
		{"length 5", 5},
		{"length 10", 10},
		{"length 19", 19},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := GenerateRandomNumber(tt.length)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			fmt.Println(n)
			s := fmt.Sprintf("%d", n)
			if len(s) != int(tt.length) {
				t.Errorf("expected length %d, got %d", tt.length, len(s))
			}
			if s[0] == '0' {
				t.Errorf("number starts with zero: %s", s)
			}
		})
	}

	// Test error cases
	t.Run("length 0", func(t *testing.T) {
		_, err := GenerateRandomNumber(0)
		if err == nil {
			t.Error("expected error for length 0, got nil")
		}
	})
	t.Run("length too large", func(t *testing.T) {
		_, err := GenerateRandomNumber(20)
		if err == nil {
			t.Error("expected error for length > 19, got nil")
		}
	})
}
