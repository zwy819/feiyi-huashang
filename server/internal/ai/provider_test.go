package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOpenAIProviderStreamsChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %q", got)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintln(w, `data: {"choices":[{"delta":{"content":"根据馆藏，"}}]}`)
		fmt.Fprintln(w)
		fmt.Fprintln(w, `data: {"choices":[{"delta":{"content":"裙褶象征迁徙记忆[S1]。"}}],"usage":{"prompt_tokens":12,"completion_tokens":8,"total_tokens":20}}`)
		fmt.Fprintln(w)
		fmt.Fprintln(w, "data: [DONE]")
	}))
	defer server.Close()

	provider := NewOpenAIProvider(server.URL, "test-key", "qwen-plus", "text-embedding-v3", 10*time.Second)
	var deltas []string
	text, usage, err := provider.ChatStream(context.Background(), []Message{{Role: "user", Content: "test"}}, func(delta string) error {
		deltas = append(deltas, delta)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(deltas, "") != text || !strings.Contains(text, "裙褶") {
		t.Fatalf("unexpected stream: deltas=%v text=%q", deltas, text)
	}
	if usage.TotalTokens != 20 {
		t.Fatalf("unexpected usage: %+v", usage)
	}
}

func TestOpenAIProviderEmbeddings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"index":0,"embedding":[0.1,0.2]},{"index":1,"embedding":[0.3,0.4]}]}`)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(server.URL, "test-key", "qwen-plus", "text-embedding-v3", 10*time.Second)
	vectors, err := provider.Embed(context.Background(), []string{"苗族", "藏袍"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vectors) != 2 || vectors[1][1] != float32(0.4) {
		t.Fatalf("unexpected vectors: %v", vectors)
	}
}

func TestOpenAIProviderReportsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":{"message":"invalid api key"}}`)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(server.URL, "bad-key", "qwen-plus", "text-embedding-v3", 10*time.Second)
	_, _, err := provider.ChatStream(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid api key") {
		t.Fatalf("unexpected error: %v", err)
	}
}
