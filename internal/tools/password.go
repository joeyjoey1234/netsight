package tools

import (
	"crypto/rand"
	"math/big"
)

func GeneratePassword(length int, includeUpper, includeDigits, includeSpecial bool) string {
	lower := "abcdefghijklmnopqrstuvwxyz"
	upper := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits := "0123456789"
	special := "!@#$%^&*()-_=+[]{}|;:,.<>?"

	charset := lower
	if includeUpper {
		charset += upper
	}
	if includeDigits {
		charset += digits
	}
	if includeSpecial {
		charset += special
	}

	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[n.Int64()]
	}
	return string(result)
}

func GeneratePSK() string {
	return GeneratePassword(63, true, true, true)
}
