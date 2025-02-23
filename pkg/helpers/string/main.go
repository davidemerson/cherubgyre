package stringhelpers

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

func GetRandomString(length int) (string, error) {
	result := make([]byte, length)
	for i := range result {
		// Generate a random index for charset
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}

func NullableStringPointer(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
