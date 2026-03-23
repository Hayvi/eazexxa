package models

import "testing"

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email   string
		wantErr bool
	}{
		{"test@example.com", false},
		{"user.name+tag@example.co.uk", false},
		{"invalid", true},
		{"@example.com", true},
		{"test@", true},
		{"", true},
	}

	for _, tt := range tests {
		err := ValidateEmail(tt.email)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
		}
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		username string
		wantErr  bool
	}{
		{"user123", false},
		{"test_user", false},
		{"ab", true},
		{"abcdefghijklmnopqrstu", true},
		{"user-name", true},
		{"", true},
	}

	for _, tt := range tests {
		err := ValidateUsername(tt.username)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateUsername(%q) error = %v, wantErr %v", tt.username, err, tt.wantErr)
		}
	}
}

func TestValidateRole(t *testing.T) {
	tests := []struct {
		role    string
		wantErr bool
	}{
		{RoleUser, false},
		{RoleAdmin, false},
		{RoleSuperAdmin, false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		err := ValidateRole(tt.role)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateRole(%q) error = %v, wantErr %v", tt.role, err, tt.wantErr)
		}
	}
}
