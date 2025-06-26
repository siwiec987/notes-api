package validation

import (
	"testing"

	"github.com/siwiec987/notes-api/internal/models"
)

func TestValidateInput_Username(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		expectedErrors []string
	}{
		{
			name:           "valid username",
			username:       "validuser123",
			expectedErrors: nil,
		},
		{
			name:           "too short",
			username:       "ab",
			expectedErrors: []string{"must be between 3 and 20 characters long"},
		},
		{
			name:           "too long",
			username:       "thisusernameistoolong123",
			expectedErrors: []string{"must be between 3 and 20 characters long"},
		},
		{
			name:           "invalid characters",
			username:       "user@name",
			expectedErrors: []string{"can only contain letters, digits, dots and underscores"},
		},
		{
			name:           "starts with dot",
			username:       ".username",
			expectedErrors: []string{"cannot start with '.' or '_'"},
		},
		{
			name:           "starts with underscore",
			username:       "_username",
			expectedErrors: []string{"cannot start with '.' or '_'"},
		},
		{
			name:           "ends with dot",
			username:       "username.",
			expectedErrors: []string{"cannot end with '.' or '_'"},
		},
		{
			name:           "ends with underscore",
			username:       "username_",
			expectedErrors: []string{"cannot end with '.' or '_'"},
		},
		{
			name:     "multiple errors",
			username: ".@",
			expectedErrors: []string{
				"must be between 3 and 20 characters long",
				"can only contain letters, digits, dots and underscores",
				"cannot start with '.' or '_'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := models.RegisterRequest{
				Username: tt.username,
				Email:    "valid@email.com", // Valid email to isolate username testing
				Password: "ValidPass123!",   // Valid password to isolate username testing
			}

			result := ValidateInput(request)
			usernameErrors := result["username"]

			if tt.expectedErrors == nil {
				if len(usernameErrors) != 0 {
					t.Errorf("Expected no username errors, got %v", usernameErrors)
				}
				return
			}

			if len(usernameErrors) != len(tt.expectedErrors) {
				t.Errorf("Expected %d username errors, got %d: %v", len(tt.expectedErrors), len(usernameErrors), usernameErrors)
				return
			}

			for _, expectedError := range tt.expectedErrors {
				found := false
				for _, actualError := range usernameErrors {
					if actualError == expectedError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error '%s' not found in %v", expectedError, usernameErrors)
				}
			}
		})
	}
}

func TestValidateInput_Email(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		expectedErrors []string
	}{
		{
			name:           "valid email",
			email:          "user@example.com",
			expectedErrors: nil,
		},
		{
			name:           "valid email with subdomain",
			email:          "user@mail.example.com",
			expectedErrors: nil,
		},
		{
			name:           "valid email with numbers",
			email:          "user123@example123.com",
			expectedErrors: nil,
		},
		{
			name:           "invalid - no @",
			email:          "userexample.com",
			expectedErrors: []string{"invalid format, example: name@example.com"},
		},
		{
			name:           "invalid - no domain",
			email:          "user@",
			expectedErrors: []string{"invalid format, example: name@example.com"},
		},
		{
			name:           "invalid - no extension",
			email:          "user@example",
			expectedErrors: []string{"invalid format, example: name@example.com"},
		},
		{
			name:           "empty email",
			email:          "",
			expectedErrors: []string{"invalid format, example: name@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := models.RegisterRequest{
				Username: "validuser",       // Valid username to isolate email testing
				Email:    tt.email,
				Password: "ValidPass123!",   // Valid password to isolate email testing
			}

			result := ValidateInput(request)
			emailErrors := result["email"]

			if tt.expectedErrors == nil {
				if len(emailErrors) != 0 {
					t.Errorf("Expected no email errors, got %v", emailErrors)
				}
				return
			}

			if len(emailErrors) != len(tt.expectedErrors) {
				t.Errorf("Expected %d email errors, got %d: %v", len(tt.expectedErrors), len(emailErrors), emailErrors)
				return
			}

			for _, expectedError := range tt.expectedErrors {
				found := false
				for _, actualError := range emailErrors {
					if actualError == expectedError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error '%s' not found in %v", expectedError, emailErrors)
				}
			}
		})
	}
}

func TestValidateInput_Password(t *testing.T) {
	tests := []struct {
		name           string
		password       string
		expectedErrors []string
	}{
		{
			name:           "valid password",
			password:       "ValidPass123!",
			expectedErrors: nil,
		},
		{
			name:           "too short",
			password:       "Short1!",
			expectedErrors: []string{"at least 8 characters long"},
		},
		{
			name:           "no uppercase",
			password:       "lowercase123!",
			expectedErrors: []string{"at least one uppercase letter"},
		},
		{
			name:           "no lowercase",
			password:       "UPPERCASE123!",
			expectedErrors: []string{"at least one lowercase letter"},
		},
		{
			name:           "no digit",
			password:       "NoDigitPass!",
			expectedErrors: []string{"at least one digit"},
		},
		{
			name:           "no special character",
			password:       "NoSpecialPass123",
			expectedErrors: []string{"at least one special character"},
		},
		{
			name:     "multiple errors",
			password: "short",
			expectedErrors: []string{
				"at least 8 characters long",
				"at least one uppercase letter",
				"at least one digit",
				"at least one special character",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := models.RegisterRequest{
				Username: "validuser",     // Valid username to isolate password testing
				Email:    "valid@email.com", // Valid email to isolate password testing
				Password: tt.password,
			}

			result := ValidateInput(request)
			passwordErrors := result["password"]

			if tt.expectedErrors == nil {
				if len(passwordErrors) != 0 {
					t.Errorf("Expected no password errors, got %v", passwordErrors)
				}
				return
			}

			if len(passwordErrors) != len(tt.expectedErrors) {
				t.Errorf("Expected %d password errors, got %d: %v", len(tt.expectedErrors), len(passwordErrors), passwordErrors)
				return
			}

			for _, expectedError := range tt.expectedErrors {
				found := false
				for _, actualError := range passwordErrors {
					if actualError == expectedError {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error '%s' not found in %v", expectedError, passwordErrors)
				}
			}
		})
	}
}

func TestValidateInput_AllValid(t *testing.T) {
	request := models.RegisterRequest{
		Username: "validuser123",
		Email:    "user@example.com",
		Password: "ValidPass123!",
	}

	result := ValidateInput(request)

	if len(result) != 0 {
		t.Errorf("Expected no validation errors for valid input, got %v", result)
	}
}

func TestValidateInput_AllInvalid(t *testing.T) {
	request := models.RegisterRequest{
		Username: "a",              // Too short
		Email:    "invalid-email",  // Invalid format
		Password: "weak",           // Multiple issues
	}

	result := ValidateInput(request)

	if len(result) != 3 {
		t.Errorf("Expected errors for all 3 fields, got errors for %d fields: %v", len(result), result)
	}

	if len(result["username"]) == 0 {
		t.Error("Expected username errors, got none")
	}
	if len(result["email"]) == 0 {
		t.Error("Expected email errors, got none")
	}
	if len(result["password"]) == 0 {
		t.Error("Expected password errors, got none")
	}
}