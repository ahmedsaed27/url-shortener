package shorturl

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("code length must be greater than zero")
	}

	code := make([]byte, length)
	max := big.NewInt(int64(len(base62Chars)))

	for i := range code {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("generate random code: %w", err)
		}

		code[i] = base62Chars[n.Int64()]
	}

	return string(code), nil
}
