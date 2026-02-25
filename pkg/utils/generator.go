package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

// GenerateEmployeeCode generates a random employee code with a prefix
// format: [PREFIX][6 DIGITS] e.g. HR123456, MA999999
func GenerateEmployeeCode(prefix string) (string, error) {
	if prefix == "" {
		prefix = "EMP"
	}
	prefix = strings.ToUpper(prefix)
	if len(prefix) > 2 {
		prefix = prefix[:2]
	}

	// Generate 6 random digits
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	// Format as 6 digits with leading zeros
	code := fmt.Sprintf("%06d", n)

	return prefix + code, nil
}
