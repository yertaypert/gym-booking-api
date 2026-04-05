package usecase

import "testing"

func TestValidateRegistrationInput(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		fullName string
		wantErr  error
	}{
		{
			name:     "valid input",
			email:    "user@example.com",
			password: "StrongPass1",
			fullName: "User Name",
		},
		{
			name:     "invalid email",
			email:    "not-an-email",
			password: "StrongPass1",
			fullName: "User Name",
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "empty full name",
			email:    "user@example.com",
			password: "StrongPass1",
			fullName: "   ",
			wantErr:  ErrInvalidFullName,
		},
		{
			name:     "weak password",
			email:    "user@example.com",
			password: "weakpass",
			fullName: "User Name",
			wantErr:  ErrWeakPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegistrationInput(tt.email, tt.password, tt.fullName)
			if err != tt.wantErr {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	got := normalizeEmail("  USER@Example.com ")
	if got != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", got)
	}
}
