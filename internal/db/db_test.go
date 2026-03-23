package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/betpro/server/internal/config"
	"github.com/jackc/pgx/v5"
)

func TestNew(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database integration test")
	}

	cfg := &config.Config{
		DBHost:             "/var/run/postgresql",
		DBPort:             5432,
		DBName:             "betpro",
		DBUser:             "postgres",
		DBPassword:         "",
		DBPoolMax:          5,
		DBPoolMin:          1,
		DBPoolIdleTimeout:  30 * time.Second,
		DBPoolConnTimeout:  5 * time.Second,
		DBQueryTimeout:     15 * time.Second,
		DBStatementTimeout: 15 * time.Second,
		DBIdleInTxTimeout:  10 * time.Second,
		DBAppName:          "betpro-test",
	}

	ctx := context.Background()
	db, err := New(ctx, cfg)
	if err != nil {
		t.Skipf("database not available: %v", err)
	}
	defer db.Close()

	if err := db.Health(ctx); err != nil {
		t.Errorf("Health() failed: %v", err)
	}

	stats := db.Stats()
	if stats.TotalConns() == 0 {
		t.Error("expected at least one connection")
	}
}

func TestWithTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database integration test")
	}

	cfg := &config.Config{
		DBHost:             "/var/run/postgresql",
		DBPort:             5432,
		DBName:             "betpro",
		DBUser:             "postgres",
		DBPoolMax:          5,
		DBPoolMin:          1,
		DBPoolIdleTimeout:  30 * time.Second,
		DBPoolConnTimeout:  5 * time.Second,
		DBQueryTimeout:     15 * time.Second,
		DBStatementTimeout: 15 * time.Second,
		DBIdleInTxTimeout:  10 * time.Second,
		DBAppName:          "betpro-test",
	}

	ctx := context.Background()
	db, err := New(ctx, cfg)
	if err != nil {
		t.Skipf("database not available: %v", err)
	}
	defer db.Close()

	t.Run("commit on success", func(t *testing.T) {
		err := db.WithTransaction(ctx, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, "SELECT 1")
			return err
		})
		if err != nil {
			t.Errorf("WithTransaction() failed: %v", err)
		}
	})

	t.Run("rollback on error", func(t *testing.T) {
		testErr := errors.New("test error")
		err := db.WithTransaction(ctx, func(tx pgx.Tx) error {
			return testErr
		})
		if err != testErr {
			t.Errorf("expected error %v, got %v", testErr, err)
		}
	})
}

func TestBuildConnString(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.Config
		expected string
	}{
		{
			name: "with password",
			cfg: &config.Config{
				DBHost:     "localhost",
				DBPort:     5432,
				DBName:     "testdb",
				DBUser:     "testuser",
				DBPassword: "testpass",
			},
			expected: "host=localhost port=5432 dbname=testdb user=testuser password=testpass sslmode=disable",
		},
		{
			name: "without password",
			cfg: &config.Config{
				DBHost: "/var/run/postgresql",
				DBPort: 5432,
				DBName: "testdb",
				DBUser: "postgres",
			},
			expected: "host=/var/run/postgresql port=5432 dbname=testdb user=postgres sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildConnString(tt.cfg)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
