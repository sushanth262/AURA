package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sushanth262/AURA/aura-backend/internal/config"
	"github.com/sushanth262/AURA/aura-backend/internal/orchestration"
)

const defaultCheckpointKeyPrefix = "aura:orch:checkpoint:"

// NewCheckpointStore selects memory or Redis backend from config.
func NewCheckpointStore(cfg config.Config) (CheckpointStore, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.CheckpointBackend)) {
	case "redis":
		return NewRedisCheckpointStore(cfg.RedisURL, cfg.CheckpointKeyPrefix)
	default:
		return NewMemoryCheckpointStore(), nil
	}
}

// RedisCheckpointStore persists GraphCheckpoint JSON in Redis.
type RedisCheckpointStore struct {
	client *redis.Client
	prefix string
}

// NewRedisCheckpointStore connects to RedisURL (e.g. redis://127.0.0.1:6379).
func NewRedisCheckpointStore(redisURL, keyPrefix string) (*RedisCheckpointStore, error) {
	redisURL = strings.TrimSpace(redisURL)
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL required when CHECKPOINT_BACKEND=redis")
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse REDIS_URL: %w", err)
	}
	if keyPrefix == "" {
		keyPrefix = defaultCheckpointKeyPrefix
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &RedisCheckpointStore{client: client, prefix: keyPrefix}, nil
}

func (s *RedisCheckpointStore) key(taskID string) string {
	return s.prefix + taskID
}

func (s *RedisCheckpointStore) Load(taskID string) (orchestration.GraphCheckpoint, bool) {
	ctx := context.Background()
	b, err := s.client.Get(ctx, s.key(taskID)).Bytes()
	if err == redis.Nil {
		return orchestration.GraphCheckpoint{}, false
	}
	if err != nil {
		return orchestration.GraphCheckpoint{}, false
	}
	var cp orchestration.GraphCheckpoint
	if err := json.Unmarshal(b, &cp); err != nil {
		return orchestration.GraphCheckpoint{}, false
	}
	return cp, true
}

func (s *RedisCheckpointStore) Save(cp orchestration.GraphCheckpoint) {
	cp.LastUpdated = time.Now().UTC()
	b, err := json.Marshal(cp)
	if err != nil {
		return
	}
	ctx := context.Background()
	_ = s.client.Set(ctx, s.key(cp.TaskID), b, 24*time.Hour).Err()
}

func (s *RedisCheckpointStore) MarkNodeComplete(taskID, nodeID string, status orchestration.InvestigationStatus) {
	cp, ok := s.Load(taskID)
	if !ok {
		cp.TaskID = taskID
	}
	cp.CurrentStatus = status
	cp.CompletedNodes = appendUnique(cp.CompletedNodes, nodeID)
	s.Save(cp)
}

func (s *RedisCheckpointStore) MarkNodeFailed(taskID, nodeID string) {
	cp, ok := s.Load(taskID)
	if !ok {
		cp.TaskID = taskID
	}
	cp.FailedNodes = appendUnique(cp.FailedNodes, nodeID)
	s.Save(cp)
}

func (s *RedisCheckpointStore) SaveAgentResult(taskID string, domain orchestration.AgentDomain, result orchestration.AgentResult) {
	cp, ok := s.Load(taskID)
	if !ok {
		cp.TaskID = taskID
	}
	if cp.AgentResults == nil {
		cp.AgentResults = make(map[string]orchestration.AgentResult)
	}
	cp.AgentResults[string(domain)] = result
	s.Save(cp)
}
