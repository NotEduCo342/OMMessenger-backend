package validation

import (
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func ValidateEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

func NormalizeUsername(username string) string {
	return strings.TrimSpace(username)
}

func ValidateUsername(username string) bool {
	username = NormalizeUsername(username)
	return usernameRe.MatchString(username)
}

func PasswordMinLength() int {
	minStr := os.Getenv("PASSWORD_MIN_LENGTH")
	if minStr == "" {
		return 10
	}
	min, err := strconv.Atoi(minStr)
	if err != nil || min < 8 {
		return 10
	}
	return min
}

func ValidatePassword(password string) bool {
	return len(password) >= PasswordMinLength()
}

func MaxMessageLength() int {
	maxStr := os.Getenv("MAX_MESSAGE_LENGTH")
	if maxStr == "" {
		return 4000
	}
	max, err := strconv.Atoi(maxStr)
	if err != nil || max < 1 {
		return 4000
	}
	return max
}

func TrimAndLimit(s string, max int) string {
	s = strings.TrimSpace(s)
	if max > 0 && len(s) > max {
		return s[:max]
	}
	return s
}
