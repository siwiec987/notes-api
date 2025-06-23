package validation

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/siwiec987/notes-api/internal/models"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._]{3,20}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

func validateUsername(username string) []string {
	var errs []string

	if len(username) < 3 || len(username) > 20 {
		errs = append(errs, "must be between 3 and 20 characters long")
	}
	if !usernameRegex.MatchString(username) {
		errs = append(errs, "can only contain letters, digits, dots and underscores")
	}
	if strings.HasPrefix(username, ".") || strings.HasPrefix(username, "_") {
		errs = append(errs, "cannot start with '.' or '_'")
	}
	if strings.HasSuffix(username, ".") || strings.HasSuffix(username, "_") {
		errs = append(errs, "cannot end with '.' or '_'")
	}

	return errs
}

func validateEmail(email string) []string {
	var errs []string

	if !emailRegex.MatchString(email) {
		errs = append(errs, "invalid format, example: name@example.com")
	}

	return errs
}

func validatePassword(password string) []string {
	var errs []string

	if len(password) < 8 {
		errs = append(errs, "at least 8 characters long")
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}
	if !hasUpper {
		errs = append(errs, "at least one uppercase letter")
	}
	if !hasLower {
		errs = append(errs, "at least one lowercase letter")
	}
	if !hasDigit {
		errs = append(errs, "at least one digit")
	}
	if !hasSpecial {
		errs = append(errs, "at least one special character")
	}

	return errs
}

func ValidateInput(u models.RegisterRequest) map[string][]string {
	errs := make(map[string][]string)

	if usernameErrs := validateUsername(u.Username); len(usernameErrs) > 0 {
		errs["username"] = usernameErrs
	}
	if emailErrs := validateEmail(u.Email); len(emailErrs) > 0 {
		errs["email"] = emailErrs
	}
	if passwordErrs := validatePassword(u.Password); len(passwordErrs) > 0 {
		errs["password"] = passwordErrs
	}

	return errs
}
