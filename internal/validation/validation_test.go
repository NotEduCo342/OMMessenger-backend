package validation

import (
	"os"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"Valid email", "user@example.com", true},
		{"Valid email with subdomain", "user@mail.example.com", true},
		{"Empty email", "", false},
		{"Email without @", "userexample.com", false},
		{"Email without domain", "user@", false},
		{"Email with spaces", "user @example.com", false},
		{"Valid email with numbers", "user123@example.com", true},
		{"Valid email with dots", "user.name@example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateEmail(tt.email)
			if result != tt.expected {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, result, tt.expected)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{"Email with uppercase", "User@EXAMPLE.COM", "user@example.com"},
		{"Email with spaces", "  user@example.com  ", "user@example.com"},
		{"Email with spaces and uppercase", "  USER@EXAMPLE.COM  ", "user@example.com"},
		{"Lowercase email", "user@example.com", "user@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeEmail(tt.email)
			if result != tt.expected {
				t.Errorf("NormalizeEmail(%q) = %q, want %q", tt.email, result, tt.expected)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected bool
	}{
		{"Valid username", "john_doe", true},
		{"Valid username with numbers", "user123", true},
		{"Valid username minimum length", "abc", true},
		{"Valid username maximum length", "a1234567890123456789012345678901", true},
		{"Username too short", "ab", false},
		{"Username too long", "a12345678901234567890123456789012", false},
		{"Username with spaces", "john doe", false},
		{"Username with special chars", "john-doe", false},
		{"Username with uppercase", "JohnDoe", true},
		{"Empty username", "", false},
		{"Username with only numbers", "12345", true},
		{"Username with only underscores", "____", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateUsername(tt.username)
			if result != tt.expected {
				t.Errorf("ValidateUsername(%q) = %v, want %v", tt.username, result, tt.expected)
			}
		})
	}
}

func TestNormalizeUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected string
	}{
		{"Username with spaces", "  john_doe  ", "john_doe"},
		{"Username no spaces", "john_doe", "john_doe"},
		{"Username with leading space", "  john_doe", "john_doe"},
		{"Username with trailing space", "john_doe  ", "john_doe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeUsername(tt.username)
			if result != tt.expected {
				t.Errorf("NormalizeUsername(%q) = %q, want %q", tt.username, result, tt.expected)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	// Set default password minimum length
	os.Unsetenv("PASSWORD_MIN_LENGTH")

	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"Valid password", "mysecurepassword123", true},
		{"Valid password minimum length", "1234567890", true},
		{"Password too short", "short", false},
		{"Password with special chars", "p@ssw0rd!123", true},
		{"Password with spaces", "my secure password", true},
		{"Empty password", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePassword(tt.password)
			if result != tt.expected {
				t.Errorf("ValidatePassword(%q) = %v, want %v", tt.password, result, tt.expected)
			}
		})
	}
}

func TestPasswordMinLength(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		expected    int
		shouldUnset bool
	}{
		{"Default minimum length", "", 10, true},
		{"Custom minimum length", "12", 12, false},
		{"Invalid env value", "invalid", 10, false},
		{"Below default minimum", "5", 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldUnset {
				os.Unsetenv("PASSWORD_MIN_LENGTH")
			} else {
				os.Setenv("PASSWORD_MIN_LENGTH", tt.envValue)
			}

			result := PasswordMinLength()
			if result != tt.expected {
				t.Errorf("PasswordMinLength() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestTrimAndLimit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		limit    int
		expected string
	}{
		{"Normal string", "hello world", 20, "hello world"},
		{"String with spaces", "  hello world  ", 20, "hello world"},
		{"String exceeding limit", "hello world this is too long", 10, "hello worl"},
		{"Empty string", "", 20, ""},
		{"String at limit", "hello", 5, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TrimAndLimit(tt.input, tt.limit)
			if result != tt.expected {
				t.Errorf("TrimAndLimit(%q, %d) = %q, want %q", tt.input, tt.limit, result, tt.expected)
			}
		})
	}
}
