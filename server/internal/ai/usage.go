package ai

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

// Limiter is a tiny per-process sliding-window limiter. A single Go instance is
// the deployment target; for horizontal scaling this should move to Redis.
type Limiter struct {
	mu     sync.Mutex
	hits   map[string][]time.Time
	perMin int
}

func NewLimiter(perMin int) *Limiter {
	return &Limiter{hits: make(map[string][]time.Time), perMin: perMin}
}

func (l *Limiter) Allow(key string) bool {
	if l == nil || l.perMin <= 0 {
		return true
	}
	now := time.Now()
	cutoff := now.Add(-time.Minute)
	l.mu.Lock()
	defer l.mu.Unlock()
	list := l.hits[key]
	kept := list[:0]
	for _, t := range list {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= l.perMin {
		l.hits[key] = kept
		return false
	}
	l.hits[key] = append(kept, now)
	return true
}

type UsageStore struct {
	db *sql.DB
}

func NewUsageStore(db *sql.DB) *UsageStore {
	return &UsageStore{db: db}
}

func (s *UsageStore) DailyTokens(ctx context.Context) (int64, error) {
	var n sql.NullInt64
	err := s.db.QueryRowContext(ctx, `
		SELECT SUM(prompt_tokens + completion_tokens) FROM ai_usage
		WHERE date(created_at) = date('now','localtime')`).Scan(&n)
	if err != nil {
		return 0, err
	}
	return n.Int64, nil
}

func (s *UsageStore) Record(ctx context.Context, record UsageRecord) error {
	status := record.Status
	if status == "" {
		status = "ok"
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO ai_usage
			(session_hash, endpoint, model, prompt_tokens, completion_tokens, latency_ms, tool_calls, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		record.SessionHash, record.Endpoint, record.Model,
		record.PromptTokens, record.CompletionTokens, record.LatencyMS,
		record.ToolCalls, status)
	return err
}

func (s *UsageStore) SaveFeedback(ctx context.Context, messageID string, rating int, reason, model string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO ai_feedback (message_id, rating, reason, model, retrieval_version)
		VALUES (?, ?, ?, ?, ?)`,
		messageID, rating, reason, model, retrievalVersion)
	return err
}
