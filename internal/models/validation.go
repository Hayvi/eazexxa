package models

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrInvalidUsername = errors.New("invalid username format")
	ErrInvalidStake    = errors.New("invalid stake amount")
	ErrInvalidOdds     = errors.New("invalid odds value")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

func ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

func ValidateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return ErrInvalidUsername
	}
	return nil
}

func ValidateRole(role string) error {
	switch role {
	case RoleUser, RoleAdmin, RoleSuperAdmin:
		return nil
	default:
		return fmt.Errorf("invalid role: %s", role)
	}
}
