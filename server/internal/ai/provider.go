package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ProviderError struct {
	Status  int
	Message string
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("model provider returned %d: %s", e.Status, e.Message)
}

// OpenAIProvider talks to any OpenAI-compatible endpoint (Qwen/DashScope,
// DeepSeek, Moonshot, vLLM, etc.). The key is held only in memory.
type OpenAIProvider struct {
	baseURL        string
	apiKey         string
	chatModel      string
	embeddingModel string
	httpClient     *http.Client
}

func NewOpenAIProvider(baseURL, apiKey, chatModel, embeddingModel string, timeout time.Duration) *OpenAIProvider {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &OpenAIProvider{
		baseURL:        strings.TrimRight(baseURL, "/"),
		apiKey:         apiKey,
		chatModel:      chatModel,
		embeddingModel: embeddingModel,
		httpClient:     &http.Client{Timeout: timeout},
	}
}

func (p *OpenAIProvider) ChatModel() string      { return p.chatModel }
func (p *OpenAIProvider) EmbeddingModel() string { return p.embeddingModel }

func (p *OpenAIProvider) ChatStream(
	ctx context.Context,
	messages []Message,
	onDelta func(string) error,
) (string, Usage, error) {
	payload := map[string]any{
		"model":       p.chatModel,
		"messages":    messages,
		"stream":      true,
		"temperature": 0.3,
		"max_tokens":  900,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", Usage{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", Usage{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", Usage{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", Usage{}, providerHTTPError(resp)
	}

	var full strings.Builder
	var usage Usage
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") || strings.HasPrefix(line, "event:") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		var event struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
			Usage *struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		if event.Usage != nil {
			usage.PromptTokens = event.Usage.PromptTokens
			usage.CompletionTokens = event.Usage.CompletionTokens
			usage.TotalTokens = event.Usage.TotalTokens
		}
		for _, choice := range event.Choices {
			delta := choice.Delta.Content
			if delta == "" {
				delta = choice.Message.Content
			}
			if delta == "" {
				continue
			}
			full.WriteString(delta)
			if onDelta != nil {
				if err := onDelta(delta); err != nil {
					return full.String(), usage, err
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return full.String(), usage, err
	}
	return full.String(), usage, nil
}

func (p *OpenAIProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	payload := map[string]any{
		"model": p.embeddingModel,
		"input": texts,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, providerHTTPError(resp)
	}
	var result struct {
		Data []struct {
			Index     int       `json:"index"`
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 8*1024*1024)).Decode(&result); err != nil {
		return nil, err
	}
	out := make([][]float32, len(texts))
	for _, item := range result.Data {
		if item.Index >= 0 && item.Index < len(out) {
			out[item.Index] = item.Embedding
		}
	}
	for i := range out {
		if len(out[i]) == 0 {
			return nil, fmt.Errorf("embedding response missing index %d", i)
		}
	}
	return out, nil
}

func providerHTTPError(resp *http.Response) error {
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	msg := strings.TrimSpace(string(raw))
	var parsed struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(raw, &parsed) == nil && parsed.Error.Message != "" {
		msg = parsed.Error.Message
	}
	if len([]rune(msg)) > 300 {
		msg = string([]rune(msg)[:300])
	}
	return &ProviderError{Status: resp.StatusCode, Message: msg}
}
