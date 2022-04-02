package core

import (
	"crypto/rand"
	"encoding/hex"
	goalone "github.com/bwmarrin/go-alone"
	"strings"
	"time"
)

var (
	sessionSigner *goalone.Sword
	sessionSalt   = []byte("session")
)

func init() {
	sessionSecret := make([]byte, 50)
	if _, err := rand.Read(sessionSecret); err != nil {
		panic(err)
	}
	sessionSigner = goalone.New(sessionSecret, goalone.Timestamp)
}

func GenerateSessionToken(userID string) string {
	combined := append([]byte(userID), sessionSalt...)
	return hex.EncodeToString(sessionSigner.Sign(combined))
}

func ValidateSessionToken(input string) (string, time.Time, error) {
	signed, err := hex.DecodeString(input)
	if err != nil {
		return "", time.Time{}, err
	}

	if _, err := sessionSigner.Unsign(signed); err != nil {
		return "", time.Time{}, AsUserError(400, err)
	}

	parsed := sessionSigner.Parse(signed)
	return strings.TrimSuffix(string(parsed.Payload), string(sessionSalt)), parsed.Timestamp, nil
}
