package core

import (
	"crypto/rand"
)

func generateRandomData(n int) []byte {
	bytes := make([]byte, n)
	_, _ = rand.Read(bytes)
	return bytes
}

func combineStringAndSalt(password string, salt []byte) []byte {
	return append([]byte(password), salt...)
}
