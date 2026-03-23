package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("SWARM_WS_URL", "ws://test:9999")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("SWARM_WS_URL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.JWTSecret != "test-secret" {
		t.Errorf("expected JWT_SECRET=test-secret, got %s", cfg.JWTSecret)
	}
	if cfg.Port != "3001" {
		t.Errorf("expected default Port=3001, got %s", cfg.Port)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				JWTSecret:  "secret",
				SwarmWSURL: "ws://test:9999",
			},
			wantErr: false,
		},
		{
			name: "missing JWT secret",
			cfg: &Config{
				SwarmWSURL: "ws://test:9999",
			},
			wantErr: true,
		},
		{
			name: "missing Swarm URL",
			cfg: &Config{
				JWTSecret: "secret",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	val := getEnvInt("TEST_INT", 10)
	if val != 42 {
		t.Errorf("expected 42, got %d", val)
	}

	val = getEnvInt("MISSING_INT", 10)
	if val != 10 {
		t.Errorf("expected fallback 10, got %d", val)
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"1", true},
		{"false", false},
		{"0", false},
		{"", false},
	}

	for _, tt := range tests {
		os.Setenv("TEST_BOOL", tt.value)
		val := getEnvBool("TEST_BOOL", false)
		if val != tt.expected {
			t.Errorf("value=%s: expected %v, got %v", tt.value, tt.expected, val)
		}
		os.Unsetenv("TEST_BOOL")
	}
}

func TestGetEnvDuration(t *testing.T) {
	os.Setenv("TEST_DURATION", "5000")
	defer os.Unsetenv("TEST_DURATION")

	val := getEnvDuration("TEST_DURATION", 1000*time.Millisecond)
	if val != 5000*time.Millisecond {
		t.Errorf("expected 5000ms, got %v", val)
	}
}

func TestParseCORSOrigins(t *testing.T) {
	origins := parseCORSOrigins("http://localhost:5173, http://localhost:4173 ,http://example.com")
	expected := []string{"http://localhost:5173", "http://localhost:4173", "http://example.com"}

	if len(origins) != len(expected) {
		t.Fatalf("expected %d origins, got %d", len(expected), len(origins))
	}

	for i, origin := range origins {
		if origin != expected[i] {
			t.Errorf("origins[%d]: expected %s, got %s", i, expected[i], origin)
		}
	}
}
