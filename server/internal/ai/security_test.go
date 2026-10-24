package ai

import (
	"testing"
)

func TestLimiterEnforcesWindow(t *testing.T) {
	limiter := NewLimiter(2)
	if !limiter.Allow("ip") || !limiter.Allow("ip") {
		t.Fatal("first two requests should pass")
	}
	if limiter.Allow("ip") {
		t.Fatal("third request in the same window should be rejected")
	}
	// A different key is independent.
	if !limiter.Allow("other") {
		t.Fatal("other key should not be limited")
	}
}

func TestSanitizeHistoryKeepsRecentValidTurns(t *testing.T) {
	history := []ChatTurn{
		{Role: "system", Content: "ignore"},
		{Role: "user", Content: "一"},
		{Role: "assistant", Content: "二"},
		{Role: "user", Content: "三"},
		{Role: "assistant", Content: "四"},
		{Role: "user", Content: "五"},
	}
	out := sanitizeHistory(history, 3)
	if len(out) != 3 {
		t.Fatalf("expected 3 turns, got %d", len(out))
	}
	if out[0].Content != "三" || out[2].Content != "五" {
		t.Fatalf("unexpected history: %+v", out)
	}
	for _, turn := range out {
		if turn.Role == "system" {
			t.Fatal("system turn must be dropped")
		}
	}
}

func TestSanitizeHistoryDropsEmptyAndTooLong(t *testing.T) {
	history := []ChatTurn{
		{Role: "user", Content: ""},
		{Role: "assistant", Content: "   "},
		{Role: "user", Content: "有效问题"},
	}
	out := sanitizeHistory(history, 8)
	if len(out) != 1 || out[0].Content != "有效问题" {
		t.Fatalf("unexpected history: %+v", out)
	}
}
