package handlers

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"lightpanel/models"
	"strconv"
	"strings"
)

const passwordIterations = 600000

func passwordHash(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, passwordIterations, 32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", passwordIterations, hex.EncodeToString(salt), hex.EncodeToString(key)), nil
}
func passwordMatches(stored, password string) bool {
	if len(password) > 1024 {
		return false
	}
	if len(stored) == 64 {
		return subtle.ConstantTimeCompare([]byte(stored), []byte(hash(password))) == 1
	} // Legacy migration.
	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations < 100000 || iterations > 1000000 {
		return false
	}
	salt, err := hex.DecodeString(parts[2])
	if err != nil || len(salt) != 16 {
		return false
	}
	want, err := hex.DecodeString(parts[3])
	if err != nil || len(want) != 32 {
		return false
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, iterations, 32)
	return err == nil && subtle.ConstantTimeCompare(key, want) == 1
}
func credentialDigest(user models.UserConfig) string {
	return hash(user.Username + "\x00" + user.Password)
}
func accountNeedsChange(user models.UserConfig) bool {
	return user.MustChangePassword || user.Password == hash("admin")
}
