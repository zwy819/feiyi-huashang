package tts

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"nonlegacy-garment-3d/server/internal/config"
	"nonlegacy-garment-3d/server/internal/model"
)

// Service 用 Qwen-TTS 生成自然中文语音，并把 WAV 缓存在本地。
// 前端通过 /api/tts/garment/:id 与 /api/tts/annotation/:id 播放。
type Service struct {
	cfg         *config.Config
	garments    *model.GarmentStore
	annotations *model.AnnotationStore
	client      *http.Client
	cacheDir    string
}

func NewService(cfg *config.Config, garments *model.GarmentStore, annotations *model.AnnotationStore) *Service {
	cacheDir := cfg.AITTSCacheDir
	if cacheDir == "" {
		cacheDir = "./data/tts"
	}
	_ = os.MkdirAll(cacheDir, 0o755)

	return &Service{
		cfg:         cfg,
		garments:    garments,
		annotations: annotations,
		client:      &http.Client{Timeout: cfg.AIRequestTimeout},
		cacheDir:    cacheDir,
	}
}

func (s *Service) Available() bool {
	return s != nil &&
		s.cfg.AIEnabled &&
		s.cfg.AITTSEnabled &&
		strings.TrimSpace(s.cfg.AIAPIKey) != ""
}

// GarmentAudio 返回一件服饰的简介 WAV。
func (s *Service) GarmentAudio(ctx context.Context, garmentID int64) ([]byte, error) {
	if !s.Available() {
		return nil, errors.New("tts not configured")
	}
	item, err := s.garments.GetByID(garmentID)
	if err != nil {
		return nil, err
	}
	return s.synthesize(ctx, garmentNarration(item))
}

// AnnotationAudio 返回一个 3D 热点的讲解 WAV；step 用于播报“第 N 站”。
func (s *Service) AnnotationAudio(ctx context.Context, annotationID int64, step int) ([]byte, error) {
	if !s.Available() {
		return nil, errors.New("tts not configured")
	}
	item, err := s.annotations.GetByID(annotationID)
	if err != nil {
		return nil, err
	}
	prefix := ""
	if step > 0 {
		prefix = fmt.Sprintf("第%d站，", step)
	}
	text := prefix + item.Title + "。" + item.Content
	return s.synthesize(ctx, text)
}

func (s *Service) synthesize(ctx context.Context, text string) ([]byte, error) {
	text = normalizeNarration(text)
	if text == "" {
		return nil, errors.New("empty narration text")
	}

	modelName := s.cfg.AITTSModel
	if modelName == "" {
		modelName = "qwen3-tts-flash"
	}
	voice := s.cfg.AITTSVoice
	if voice == "" {
		voice = "Cherry"
	}

	cachePath := filepath.Join(s.cacheDir, cacheKey(modelName, voice, text)+".wav")
	if data, err := os.ReadFile(cachePath); err == nil && len(data) > 44 {
		return data, nil
	}

	audio, err := s.callQwen(ctx, modelName, voice, text)
	if err != nil {
		return nil, err
	}
	if len(audio) < 44 || !bytes.HasPrefix(audio, []byte("RIFF")) {
		return nil, errors.New("tts returned invalid wav")
	}

	tmp := cachePath + ".tmp"
	if err := os.WriteFile(tmp, audio, 0o644); err == nil {
		_ = os.Rename(tmp, cachePath)
	}
	return audio, nil
}

func (s *Service) callQwen(ctx context.Context, modelName, voice, text string) ([]byte, error) {
	body, err := json.Marshal(map[string]any{
		"model": modelName,
		"input": map[string]string{
			"text":          text,
			"voice":         voice,
			"language_type": "Chinese",
		},
	})
	if err != nil {
		return nil, err
	}

	endpoint := s.ttsEndpoint()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.AIAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tts http %d: %s", resp.StatusCode, string(raw[:min(len(raw), 500)]))
	}

	var parsed struct {
		Output struct {
			Audio struct {
				URL  string `json:"url"`
				Data string `json:"data"`
			} `json:"audio"`
		} `json:"output"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("tts response: %w", err)
	}
	if parsed.Code != "" || parsed.Message != "" {
		return nil, fmt.Errorf("tts error: %s %s", parsed.Code, parsed.Message)
	}

	if parsed.Output.Audio.Data != "" {
		return base64.StdEncoding.DecodeString(parsed.Output.Audio.Data)
	}
	if parsed.Output.Audio.URL == "" {
		return nil, errors.New("tts response missing audio url")
	}

	audioResp, err := s.client.Get(parsed.Output.Audio.URL)
	if err != nil {
		return nil, err
	}
	defer audioResp.Body.Close()
	if audioResp.StatusCode < 200 || audioResp.StatusCode >= 300 {
		return nil, fmt.Errorf("tts audio download http %d", audioResp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(audioResp.Body, 16<<20))
}

// ttsEndpoint 复用当前百炼工作空间的 host，避免工作空间 Key 在通用地址上 401。
func (s *Service) ttsEndpoint() string {
	base := strings.TrimRight(s.cfg.AIBaseURL, "/")
	if parsed, err := url.Parse(base); err == nil && parsed.Hostname() != "" {
		host := parsed.Hostname()
		if strings.Contains(host, "maas.aliyuncs.com") || strings.Contains(host, "maas.qwencloudapi.com") {
			return "https://" + host + "/api/v1/services/aigc/multimodal-generation/generation"
		}
	}
	return "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation"
}

func garmentNarration(item *model.Garment) string {
	dynasty := strings.TrimSpace(item.Dynasty)
	if dynasty != "" && !strings.HasSuffix(dynasty, "代") {
		dynasty += "代"
	}
	return fmt.Sprintf(
		"%s。%s，%s，%s。工艺：%s。材质：%s。纹样：%s。%s",
		item.Name,
		fallback(dynasty),
		fallback(item.Ethnic),
		fallback(item.Province),
		fallback(item.Craft),
		fallback(item.Material),
		fallback(item.Pattern),
		fallback(item.CulturalMeaning),
	)
}

func fallback(value string) string {
	if strings.TrimSpace(value) == "" {
		return "未标注"
	}
	return value
}

func normalizeNarration(text string) string {
	text = strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	runes := []rune(text)
	if len(runes) > 600 {
		text = string(runes[:600])
		if !strings.HasSuffix(text, "。") && !strings.HasSuffix(text, "！") && !strings.HasSuffix(text, "？") {
			text += "。"
		}
	}
	return text
}

func cacheKey(modelName, voice, text string) string {
	sum := sha256.Sum256([]byte(modelName + "\x00" + voice + "\x00" + text))
	return hex.EncodeToString(sum[:])[:32]
}
