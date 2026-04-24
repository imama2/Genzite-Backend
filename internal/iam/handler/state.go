package handler

import (
	"crypto/rand"
	"encoding/base64"
)

var randReader = rand.Reader

func encodeState(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}
