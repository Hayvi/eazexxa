package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/betpro/server/pkg/logger"
)

type DistributedLock struct {
	redis *RedisClient
}

func NewDistributedLock(redis *RedisClient) *DistributedLock {
	return &DistributedLock{redis: redis}
}

func (dl *DistributedLock) AcquireLeaderLock(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if !dl.redis.IsEnabled() {
		return "", fmt.Errorf("redis not enabled")
	}

	token := generateLockToken()
	
	success, err := dl.redis.Client().SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return "", err
	}

	if !success {
		return "", fmt.Errorf("lock already held")
	}

	return token, nil
}

func (dl *DistributedLock) ReleaseLeaderLock(ctx context.Context, key, token string) error {
	if !dl.redis.IsEnabled() {
		return fmt.Errorf("redis not enabled")
	}

	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := dl.redis.Client().Eval(ctx, script, []string{key}, token).Result()
	if err != nil {
		return err
	}

	if result.(int64) == 0 {
		return fmt.Errorf("lock not held or expired")
	}

	return nil
}

func (dl *DistributedLock) RefreshLeaderLock(ctx context.Context, key, token string, ttl time.Duration) error {
	if !dl.redis.IsEnabled() {
		return fmt.Errorf("redis not enabled")
	}

	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("EXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := dl.redis.Client().Eval(ctx, script, []string{key}, token, int(ttl.Seconds())).Result()
	if err != nil {
		return err
	}

	if result.(int64) == 0 {
		return fmt.Errorf("lock not held or expired")
	}

	return nil
}

func (dl *DistributedLock) AcquireGameLock(ctx context.Context, gameID string, ttl time.Duration) (string, error) {
	key := fmt.Sprintf("lock:game:%s", gameID)
	return dl.AcquireLeaderLock(ctx, key, ttl)
}

func (dl *DistributedLock) ReleaseGameLock(ctx context.Context, gameID, token string) error {
	key := fmt.Sprintf("lock:game:%s", gameID)
	return dl.ReleaseLeaderLock(ctx, key, token)
}

func (dl *DistributedLock) CheckIdempotency(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if !dl.redis.IsEnabled() {
		return false, nil
	}

	exists, err := dl.redis.Exists(ctx, key)
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

func (dl *DistributedLock) MarkIdempotency(ctx context.Context, key string, ttl time.Duration) error {
	if !dl.redis.IsEnabled() {
		return nil
	}

	return dl.redis.Set(ctx, key, "1", ttl)
}

func (dl *DistributedLock) TryBreakStaleLock(ctx context.Context, key string) (bool, error) {
	if !dl.redis.IsEnabled() {
		return false, fmt.Errorf("redis not enabled")
	}

	currentToken, err := dl.redis.Get(ctx, key)
	if err != nil {
		return false, err
	}

	parsed := parseLockToken(currentToken)
	if parsed == nil {
		return false, nil
	}

	hostname, _ := os.Hostname()
	if parsed.Host != hostname {
		return false, nil
	}

	if isPidAlive(parsed.PID) {
		return false, nil
	}

	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := dl.redis.Client().Eval(ctx, script, []string{key}, currentToken).Result()
	if err != nil {
		return false, err
	}

	if result.(int64) > 0 {
		logger.Warn("removed stale lock", "key", key, "token", currentToken)
		return true, nil
	}

	return false, nil
}

func generateLockToken() string {
	hostname, _ := os.Hostname()
	pid := os.Getpid()
	random := make([]byte, 16)
	rand.Read(random)
	uuid := hex.EncodeToString(random)
	return fmt.Sprintf("%s:%d:%s", hostname, pid, uuid)
}

type lockToken struct {
	Host string
	PID  int
	UUID string
}

func parseLockToken(token string) *lockToken {
	parts := strings.Split(token, ":")
	if len(parts) != 3 {
		return nil
	}

	pid, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil
	}

	return &lockToken{
		Host: parts[0],
		PID:  pid,
		UUID: parts[2],
	}
}

func isPidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(os.Signal(nil))
	return err == nil
}
