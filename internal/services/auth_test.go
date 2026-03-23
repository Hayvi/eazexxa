package services

import (
	"testing"

	"github.com/betpro/server/internal/config"
)

func TestGenerateToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	authService := NewAuthService(cfg)

	token, err := authService.GenerateToken("user123", "user")
	if err != nil {
		t.Fatalf("GenerateToken() failed: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestVerifyToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	authService := NewAuthService(cfg)

	token, err := authService.GenerateToken("user123", "user")
	if err != nil {
		t.Fatalf("GenerateToken() failed: %v", err)
	}

	claims, err := authService.VerifyToken(token)
	if err != nil {
		t.Fatalf("VerifyToken() failed: %v", err)
	}

	if claims.UserID != "user123" {
		t.Errorf("expected UserID=user123, got %s", claims.UserID)
	}

	if claims.Role != "user" {
		t.Errorf("expected Role=user, got %s", claims.Role)
	}
}

func TestVerifyTokenInvalid(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	authService := NewAuthService(cfg)

	_, err := authService.VerifyToken("invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestHashPassword(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	authService := NewAuthService(cfg)

	password := "testpassword123"
	hash, err := authService.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	if hash == "" {
		t.Error("expected non-empty hash")
	}

	if hash == password {
		t.Error("hash should not equal plain password")
	}
}

func TestCheckPassword(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}
	authService := NewAuthService(cfg)

	password := "testpassword123"
	hash, err := authService.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	if err := authService.CheckPassword(password, hash); err != nil {
		t.Errorf("CheckPassword() failed: %v", err)
	}

	if err := authService.CheckPassword("wrongpassword", hash); err == nil {
		t.Error("expected error for wrong password")
	}
}
