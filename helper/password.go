package helper

import (
	"os"

	"golang.org/x/crypto/bcrypt"
	"time"
)

// HashPassword membuat hash dari password plaintext
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash membandingkan password plaintext dengan hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GetEnvOrDefault membaca environment variable jika ada,
// kalau tidak ada maka menggunakan nilai default
func GetEnvOrDefault(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// ParseDurationOrDefault mengubah string menjadi time.Duration,
// jika gagal maka mengembalikan default
func ParseDurationOrDefault(str string, defaultValue time.Duration) time.Duration {
	d, err := time.ParseDuration(str)
	if err != nil || d <= 0 {
		return defaultValue
	}
	return d
}
