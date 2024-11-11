package auth

import (
	"crypto/subtle"

	"golang.org/x/crypto/scrypt"
)

// Credential stores username and hashed credentials
type Credential struct {
	Username string
	Hash     []byte // scrypt hash of username_passphrase
}

// HashCredentials creates a secure hash using scrypt
func HashCredentials(username, passphrase string) ([]byte, error) {
	// Combine username and passphrase
	combined := passphrase[3:8] + passphrase[0:3] + username + passphrase[8:]

	// scrypt parameters
	N := 32768   // CPU/memory cost parameter
	r := 8       // block size parameter
	p := 1       // parallelization parameter
	keyLen := 32 // length of the generated key

	// Generate salt from username (since it's unique)
	salt := []byte("salt_" + username)

	return scrypt.Key([]byte(combined), salt, N, r, p, keyLen)
}

// SecureCompare performs constant-time comparison of hashes
func SecureCompare(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}
