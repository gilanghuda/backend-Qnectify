package utils

import (
	"crypto/rand"
	"math/big"
)

const inviteChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func GenerateInviteCode(length int) (string, error) {
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(inviteChars))))
		if err != nil {
			return "", err
		}
		result[i] = inviteChars[num.Int64()]
	}
	return string(result), nil
}
