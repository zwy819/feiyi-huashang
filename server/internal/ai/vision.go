package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type VisualAttributes struct {
	Ethnic      string   `json:"ethnic"`
	Craft       string   `json:"craft"`
	Pattern     string   `json:"pattern"`
	Dynasty     string   `json:"dynasty"`
	Keywords    []string `json:"keywords"`
	Confidence  float64  `json:"confidence"`
	Description string   `json:"description"`
}

type VisualCandidate struct {
	GarmentID int64   `json:"garment_id"`
	Name      string  `json:"name"`
	Dynasty   string  `json:"dynasty"`
	Ethnic    string  `json:"ethnic"`
	Craft     string  `json:"craft"`
	Province  string  `json:"province"`
	ModelID   *int64  `json:"model_id,omitempty"`
	Score     float64 `json:"score"`
	Reason    string  `json:"reason"`
}

type VisualSearchResult struct {
	Attributes VisualAttributes  `json:"attributes"`
	Candidates []VisualCandidate `json:"candidates"`
	Model      string            `json:"model"`
	LatencyMS  int64             `json:"latency_ms"`
}

const visionPrompt = `你是中国非物质文化遗产服饰的图像识别助手。请观察用户上传的照片，只提取你能从画面中合理判断的信息：
民族（如苗族/藏族/朝鲜族）、主要工艺（如刺绣/蜡染/织锦/银饰）、纹样（如花卉纹/几何纹/龙凤纹）、朝代或时代（如清/现代）。
无法判断的字段返回空字符串，不要猜测；confidence 为 0 到 1 的小数。只输出严格 JSON，不要 markdown：
{"ethnic":"","craft":"","pattern":"","dynasty":"","keywords":[""],"confidence":0,"description":"一句话描述"}`

func (s *Service) VisualSearch(ctx context.Context, ip, imageDataURL string) (*VisualSearchResult, error) {
	if s == nil || !s.Available() || s.provider == nil {
		return nil, errNotConfigured
	}
	if !s.limiter.Allow(ip) {
		return nil, errRateLimited
	}
	if !strings.HasPrefix(imageDataURL, "data:image/") {
		return nil, &ChatError{Code: 40002, Message: "图片格式不正确", Retryable: false}
	}

	start := time.Now()
	raw, err := s.provider.VisionJSON(ctx, s.cfg.AIVisionModel, imageDataURL, visionPrompt)
	if err != nil {
		return nil, err
	}
	attrs, err := parseVisualAttributes(raw)
	if err != nil {
		return nil, fmt.Errorf("parse vision result: %w", err)
	}

	candidates, err := s.rankVisualCandidates(ctx, attrs)
	if err != nil {
		return nil, err
	}
	return &VisualSearchResult{
		Attributes: attrs,
		Candidates: candidates,
		Model:      s.cfg.AIVisionModel,
		LatencyMS:  time.Since(start).Milliseconds(),
	}, nil
}

func parseVisualAttributes(raw string) (VisualAttributes, error) {
	text := strings.TrimSpace(raw)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return VisualAttributes{}, fmt.Errorf("no json object: %q", text)
	}
	var attrs VisualAttributes
	if err := json.Unmarshal([]byte(text[start:end+1]), &attrs); err != nil {
		return VisualAttributes{}, err
	}
	if attrs.Confidence < 0 {
		attrs.Confidence = 0
	}
	if attrs.Confidence > 1 {
		attrs.Confidence = 1
	}
	return attrs, nil
}

func (s *Service) rankVisualCandidates(ctx context.Context, attrs VisualAttributes) ([]VisualCandidate, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, dynasty, ethnic, craft, material, pattern, province, description, model_id FROM garments`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []VisualCandidate
	for rows.Next() {
		var item VisualCandidate
		var material, pattern, description string
		var modelID sql.NullInt64
		if err := rows.Scan(&item.GarmentID, &item.Name, &item.Dynasty, &item.Ethnic, &item.Craft,
			&material, &pattern, &item.Province, &description, &modelID); err != nil {
			return nil, err
		}
		if modelID.Valid {
			item.ModelID = &modelID.Int64
		}
		score, reasons := visualScore(attrs, item, material, pattern, description)
		item.Score = score
		item.Reason = strings.Join(reasons, "；")
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].GarmentID < out[j].GarmentID
		}
		return out[i].Score > out[j].Score
	})
	if len(out) > 5 {
		out = out[:5]
	}
	return out, nil
}

func visualScore(attrs VisualAttributes, item VisualCandidate, material, pattern, description string) (float64, []string) {
	score := 0.0
	var reasons []string
	if attrs.Ethnic != "" && textMatch(item.Ethnic, attrs.Ethnic) {
		score += 4
		reasons = append(reasons, "民族匹配「"+item.Ethnic+"」")
	}
	if attrs.Craft != "" && textMatch(item.Craft, attrs.Craft) {
		score += 3
		reasons = append(reasons, "工艺匹配「"+item.Craft+"」")
	}
	if attrs.Pattern != "" && (textMatch(pattern, attrs.Pattern) || textMatch(description, attrs.Pattern)) {
		score += 2
		reasons = append(reasons, "纹样匹配「"+pattern+"」")
	}
	if attrs.Dynasty != "" && textMatch(item.Dynasty, attrs.Dynasty) {
		score += 1
		reasons = append(reasons, "朝代匹配「"+item.Dynasty+"」")
	}
	haystack := item.Name + " " + item.Craft + " " + item.Ethnic + " " + pattern + " " + material + " " + description
	for _, keyword := range attrs.Keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword != "" && textMatch(haystack, keyword) {
			score += 0.5
			if len(reasons) < 3 {
				reasons = append(reasons, "关键词「"+keyword+"」")
			}
		}
	}
	return score, reasons
}

func textMatch(haystack, needle string) bool {
	haystack = strings.ToLower(strings.TrimSpace(haystack))
	needle = strings.ToLower(strings.TrimSpace(needle))
	if haystack == "" || needle == "" {
		return false
	}
	if strings.Contains(haystack, needle) || strings.Contains(needle, haystack) {
		return true
	}
	for _, part := range strings.FieldsFunc(needle, func(r rune) bool {
		return r == '+' || r == '、' || r == '/' || r == '，' || r == ' '
	}) {
		part = strings.TrimSpace(part)
		if len([]rune(part)) >= 2 && strings.Contains(haystack, part) {
			return true
		}
	}
	return false
}
