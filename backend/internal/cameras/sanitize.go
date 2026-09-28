package cameras

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
)

var idRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

func ValidID(id string) bool { return idRe.MatchString(id) && len(id) <= 64 }

// SanitizeRTSPForLog replaces credentials with *** so secrets never leak.
func SanitizeRTSPForLog(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return "rtsp://***"
	}
	u.User = url.User("***")
	return u.String()
}

// RedactedHost returns rtsp://host:port/path without userinfo for API output.
func RedactedHost(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	u.User = nil
	return u.String()
}

// ParseHostPort extracts host + port from an RTSP url (fallback to defaults).
func ParseHostPort(raw string, defPort int) (string, int) {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return "", defPort
	}
	port := defPort
	if p := u.Port(); p != "" {
		var n int
		if _, err := fmt.Sscanf(p, "%d", &n); err == nil {
			port = n
		}
	}
	return u.Hostname(), port
}

func key32(secret string) []byte {
	h := sha256.Sum256([]byte(secret))
	return h[:]
}

// Encrypt password with AES-GCM + random nonce, base64 encoded. Never logged.
func Encrypt(secret, plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	block, err := aes.NewCipher(key32(secret))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ct), nil
}

func Decrypt(secret, encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key32(secret))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := raw[:gcm.NonceSize()]
	ct := raw[gcm.NonceSize():]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

// InjectPassword rebuilds full RTSP url from stored template + decrypted password.
func InjectPassword(storedURL, username, password string) string {
	if password == "" {
		return storedURL
	}
	u, err := url.Parse(storedURL)
	if err != nil {
		return storedURL
	}
	if username != "" {
		u.User = url.UserPassword(username, password)
	}
	return u.String()
}

// StripPassword removes userinfo for storage template (keeps host/path).
func StripPassword(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return strings.TrimSpace(raw)
	}
	u.User = nil
	return u.String()
}
