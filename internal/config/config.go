package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port         string
	NodeEnv      string
	TrustProxy   int
	CORSOrigins  []string

	// JWT
	JWTSecret    string
	JWTExpiresIn string

	// Database
	DBHost                 string
	DBPort                 int
	DBName                 string
	DBUser                 string
	DBPassword             string
	DBPoolMax              int
	DBPoolMin              int
	DBPoolIdleTimeout      time.Duration
	DBPoolConnTimeout      time.Duration
	DBQueryTimeout         time.Duration
	DBStatementTimeout     time.Duration
	DBIdleInTxTimeout      time.Duration
	DBPoolMaxUses          int
	DBAppName              string

	// Redis
	RedisURL     string
	RedisEnabled bool
	RedisWSChannel string

	// Swarm
	SwarmWSURL  string
	SwarmSiteID string
	SwarmLang   string

	// Settlement
	SettlementLeaderLockKey          string
	SettlementLeaderTTL              time.Duration
	SettlementGameLockTTL            time.Duration
	SettlementIdempotencyTTL         time.Duration
	SettlementStaleFinalOutcomeTTL   time.Duration
	SettlementNotFinalResolverTTL    time.Duration
	SettlementRedisFailOpen          bool
	SettlementPollInterval           time.Duration

	// Betting
	BetPlacementSlowThreshold time.Duration

	// Workers
	WSNodeID string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:         getEnv("PORT", "3001"),
		NodeEnv:      getEnv("NODE_ENV", "development"),
		TrustProxy:   getEnvInt("TRUST_PROXY", 1),
		CORSOrigins:  parseCORSOrigins(getEnv("CORS_ORIGIN", "http://localhost:5173,http://localhost:4173")),

		JWTSecret:    getEnv("JWT_SECRET", ""),
		JWTExpiresIn: getEnv("JWT_EXPIRES_IN", "7d"),

		DBHost:                 getEnv("DB_HOST", "/var/run/postgresql"),
		DBPort:                 getEnvInt("DB_PORT", 5432),
		DBName:                 getEnv("DB_NAME", "betpro"),
		DBUser:                 getEnv("DB_USER", "postgres"),
		DBPassword:             getEnv("DB_PASSWORD", ""),
		DBPoolMax:              getEnvInt("DB_POOL_MAX", 20),
		DBPoolMin:              getEnvInt("DB_POOL_MIN", 2),
		DBPoolIdleTimeout:      getEnvDuration("DB_POOL_IDLE_TIMEOUT_MS", 30000*time.Millisecond),
		DBPoolConnTimeout:      getEnvDuration("DB_POOL_CONN_TIMEOUT_MS", 5000*time.Millisecond),
		DBQueryTimeout:         getEnvDuration("DB_QUERY_TIMEOUT_MS", 15000*time.Millisecond),
		DBStatementTimeout:     getEnvDuration("DB_STATEMENT_TIMEOUT_MS", 15000*time.Millisecond),
		DBIdleInTxTimeout:      getEnvDuration("DB_IDLE_IN_TX_TIMEOUT_MS", 10000*time.Millisecond),
		DBPoolMaxUses:          getEnvInt("DB_POOL_MAX_USES", 7500),
		DBAppName:              getEnv("DB_APP_NAME", "betpro-server"),

		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		RedisEnabled:   getEnvBool("REDIS_ENABLED", true),
		RedisWSChannel: getEnv("REDIS_WS_CHANNEL", "betpro:ws:broadcast"),

		SwarmWSURL:  getEnv("SWARM_WS_URL", ""),
		SwarmSiteID: getEnv("SWARM_SITE_ID", "4"),
		SwarmLang:   getEnv("SWARM_LANG", "eng"),

		SettlementLeaderLockKey:        getEnv("SETTLEMENT_LEADER_LOCK_KEY", "betpro:settlement:leader"),
		SettlementLeaderTTL:            getEnvDuration("SETTLEMENT_LEADER_TTL_SEC", 900*time.Second),
		SettlementGameLockTTL:          getEnvDuration("SETTLEMENT_GAME_LOCK_TTL_SEC", 120*time.Second),
		SettlementIdempotencyTTL:       getEnvDuration("SETTLEMENT_IDEMPOTENCY_TTL_SEC", 600*time.Second),
		SettlementStaleFinalOutcomeTTL: getEnvDuration("SETTLEMENT_STALE_FINAL_OUTCOME_TTL_SEC", 900*time.Second),
		SettlementNotFinalResolverTTL:  getEnvDuration("SETTLEMENT_NOT_FINAL_RESOLVER_TTL_SEC", 172800*time.Second),
		SettlementRedisFailOpen:        getEnvBool("SETTLEMENT_REDIS_FAIL_OPEN", true),
		SettlementPollInterval:         getEnvDuration("SETTLEMENT_POLL_INTERVAL_MS", 60000*time.Millisecond),

		BetPlacementSlowThreshold: getEnvDuration("BET_PLACEMENT_SLOW_THRESHOLD_MS", 1000*time.Millisecond),

		WSNodeID: getEnv("WS_NODE_ID", fmt.Sprintf("%d", os.Getpid())),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.SwarmWSURL == "" {
		return fmt.Errorf("SWARM_WS_URL is required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true" || val == "1"
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return time.Duration(i) * time.Millisecond
		}
	}
	return fallback
}

func parseCORSOrigins(val string) []string {
	parts := strings.Split(val, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
