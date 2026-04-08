// Package crypto provides cryptographic utilities for NixGuard.
package crypto

import (
	"crypto/rand"
	"encoding/hex"
)

// RandomID generates a cryptographically secure random hex string.
func RandomID(byteLen int) string {
	b := make([]byte, byteLen)
	rand.Read(b)
	return hex.EncodeToString(b)
}
