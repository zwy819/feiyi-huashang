package ai

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nonlegacy-garment-3d/server/internal/config"
	"nonlegacy-garment-3d/server/internal/model"
)

type ChatError struct {
	Code      int
	Message   string
	Retryable bool
}

func (e *ChatError) Error() string { return e.Message }

var (
	errNotConfigured = &ChatError{Code: 50301, Message: "AI 导览暂未配置，请先设置 AI_API_KEY", Retryable: false}
	errRateLimited   = &ChatError{Code: 42901, Message: "提问太频繁了，请稍后再试", Retryable: true}
	errQuota         = &ChatError{Code: 42902, Message: "今日 AI 导览额度已用完，请明天再来", Retryable: false}
)

type EmitFunc func(event string, data any) error

type Service struct {
	cfg          *config.Config
	db           *sql.DB
	store        *KnowledgeStore
	retriever    *Retriever
	ingestor     *Ingestor
	toolbox      *Toolbox
	provider     *OpenAIProvider
	embedder     Embedder
	usage        *UsageStore
	limiter      *Limiter
	knowledgeDir string
}

func NewService(cfg *config.Config, db *sql.DB, garmentStore *model.GarmentStore, annotationStore *model.AnnotationStore, knowledgeDir string) *Service {
	knowledgeDir = resolveKnowledgeDir(knowledgeDir)
	store := NewKnowledgeStore(db)

	var provider *OpenAIProvider
	if cfg.AIAPIKey != "" && !cfg.AIMockMode {
		provider = NewOpenAIProvider(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIChatModel, cfg.AIEmbeddingModel, cfg.AIRequestTimeout)
	}
	var embedder Embedder
	if provider != nil {
		embedder = provider
	}

	return &Service{
		cfg:          cfg,
		db:           db,
		store:        store,
		retriever:    NewRetriever(store, embedder),
		ingestor:     NewIngestor(store, garmentStore, annotationStore, knowledgeDir),
		toolbox:      NewToolbox(db, garmentStore, annotationStore),
		provider:     provider,
		embedder:     embedder,
		usage:        NewUsageStore(db),
		limiter:      NewLimiter(cfg.AIRateLimitPerMin),
		knowledgeDir: knowledgeDir,
	}
}

func resolveKnowledgeDir(dir string) string {
	if dir == "" {
		dir = filepath.Join("knowledge", "articles")
	}
	candidates := []string{dir, filepath.Join("server", dir)}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return dir
}

// Init builds the corpus and the in-memory index. Failure is logged by the
// caller but never prevents the ordinary museum pages from starting.
func (s *Service) Init(ctx context.Context, force bool) error {
	stats, err := s.ingestor.Ingest(ctx, force, s.embedder)
	if err != nil {
		return err
	}
	if err := s.retriever.Refresh(ctx); err != nil {
		return err
	}
	log.Printf("[AI] 知识库就绪: documents=%d chunks=%d embedded=%d skipped=%d",
		stats.Documents, stats.Chunks, stats.Embedded, stats.Skipped)
	return nil
}

func (s *Service) Available() bool {
	return s.cfg.AIEnabled && (s.cfg.AIMockMode || s.provider != nil)
}

func (s *Service) Configured() bool {
	return s.provider != nil
}

func (s *Service) Status(ctx context.Context) Status {
	chunks, _ := s.store.CountChunks(ctx)
	status := Status{
		Enabled:        s.cfg.AIEnabled,
		Configured:     s.Configured() || s.cfg.AIMockMode,
		MockMode:       s.cfg.AIMockMode,
		Model:          s.cfg.AIChatModel,
		EmbeddingModel: s.cfg.AIEmbeddingModel,
		Chunks:         chunks,
	}
	switch {
	case !s.cfg.AIEnabled:
		status.Message = "AI 导览已关闭"
	case s.cfg.AIMockMode:
		status.Message = "演示模式：使用内置回答，不调用外部模型"
	case s.provider == nil:
		status.Message = "未配置 AI_API_KEY，普通浏览不受影响"
	default:
		status.Message = "已连接 " + s.cfg.AIChatModel
	}
	return status
}

func (s *Service) Chat(ctx context.Context, req ChatRequest, clientKey string, emit EmitFunc) error {
	start := time.Now()
	if !s.Available() {
		return errNotConfigured
	}
	message := strings.TrimSpace(req.Message)
	if message == "" || len([]rune(message)) > 500 {
		return &ChatError{Code: 40001, Message: "问题长度需在 1–500 字之间", Retryable: false}
	}
	if len(req.History) > s.cfg.AIMaxHistoryTurns*2 {
		req.History = req.History[len(req.History)-s.cfg.AIMaxHistoryTurns*2:]
	}
	if !s.limiter.Allow(clientKey) {
		return errRateLimited
	}
	if s.cfg.AIDailyTokenBudget > 0 {
		tokens, err := s.usage.DailyTokens(ctx)
		if err == nil && tokens >= s.cfg.AIDailyTokenBudget {
			return errQuota
		}
	}

	conversationID := randomID("conv_")
	messageID := randomID("msg_")
	sessionHash := hashKey(clientKey)
	if err := emit("meta", map[string]any{
		"conversation_id": conversationID,
		"message_id":      messageID,
		"model":           s.cfg.AIChatModel,
		"mock_mode":       s.cfg.AIMockMode,
		"started_at":      start.UTC().Format(time.RFC3339),
	}); err != nil {
		return err
	}

	runCtx, cancel := context.WithTimeout(ctx, s.cfg.AIRequestTimeout)
	defer cancel()

	if err := s.chat(runCtx, req, message, conversationID, sessionHash, messageID, emit); err != nil {
		latency := time.Since(start).Milliseconds()
		_ = s.usage.Record(ctx, UsageRecord{
			SessionHash: sessionHash, Endpoint: "chat", Model: s.cfg.AIChatModel,
			LatencyMS: latency, Status: statusForError(err),
		})
		var chatErr *ChatError
		if errors.As(err, &chatErr) {
			_ = emit("error", map[string]any{
				"code": chatErr.Code, "message": chatErr.Message, "retryable": chatErr.Retryable,
			})
			return nil
		}
		_ = emit("error", map[string]any{
			"code": 50201, "message": "AI 导览暂时不可用，请稍后重试", "retryable": true,
		})
		return err
	}
	return nil
}

func (s *Service) chat(ctx context.Context, req ChatRequest, message, conversationID, sessionHash, messageID string, emit EmitFunc) error {
	start := time.Now()
	toolCalls := 0
	var contextGarment *model.Garment
	var contextAnnotation *model.Annotation

	if req.Context.GarmentID > 0 {
		if g, err := s.toolbox.GetGarment(ctx, req.Context.GarmentID); err == nil {
			contextGarment = g
		}
	}
	if req.Context.AnnotationID > 0 {
		if a, err := s.toolbox.annotationStore.GetByID(req.Context.AnnotationID); err == nil {
			if contextGarment == nil || a.GarmentID == contextGarment.ID {
				contextAnnotation = a
			}
		}
	}
	history := sanitizeHistory(req.History, s.cfg.AIMaxHistoryTurns*2)

	filters, _ := s.toolbox.ExtractFilters(ctx, message)
	namedGarment, _ := s.toolbox.FindGarmentByName(ctx, message)
	queryGarment := contextGarment
	if namedGarment != nil {
		if g, err := s.toolbox.GetGarment(ctx, namedGarment.ID); err == nil {
			queryGarment = g
		}
	}

	evidence := make([]evidenceItem, 0, 10)

	if contextGarment != nil {
		evidence = append(evidence, evidenceItem{
			Source: Source{
				Type: "馆藏", Title: "馆藏档案 · " + contextGarment.Name,
				Publisher: "非遗华裳馆藏", URL: fmt.Sprintf("/detail/%d", contextGarment.ID),
				Snippet:   firstSentence(garmentDocument(*contextGarment), 160),
				GarmentID: contextGarment.ID,
			},
			Content: garmentDocument(*contextGarment),
		})
	}
	if contextAnnotation != nil {
		evidence = append(evidence, evidenceItem{
			Source: Source{
				Type: "3D 讲解", Title: "3D 讲解 · " + contextAnnotation.Title,
				Publisher: "非遗华裳 3D 讲解",
				URL:       fmt.Sprintf("/detail/%d?annotation=%d", contextAnnotation.GarmentID, contextAnnotation.ID),
				Snippet:   firstSentence(contextAnnotation.Content, 160),
				GarmentID: contextAnnotation.GarmentID, AnnotationID: contextAnnotation.ID,
			},
			Content: contextAnnotation.Content,
		})
	}

	// Statistics questions are answered from SQL aggregates, not model guesses.
	statsIntent := statsDimension(message)
	if statsIntent != "" {
		toolStart := time.Now()
		items, err := s.toolbox.GetStatistics(ctx, statsIntent)
		if err == nil {
			summary := statsSummary(statsIntent, items)
			evidence = append(evidence, evidenceItem{
				Source: Source{
					Type: "统计", Title: "馆藏统计 · " + dimensionLabel(statsIntent),
					Publisher: "非遗华裳馆藏", URL: "/statistics", Snippet: summary,
				},
				Content: summary,
			})
			toolCalls++
			if err := s.emitTool(emit, "get_statistics", fmt.Sprintf("按%s统计，返回 %d 组", dimensionLabel(statsIntent), len(items)), toolStart); err != nil {
				return err
			}
		}
	}

	if !filters.Empty() || namedGarment != nil {
		toolStart := time.Now()
		items, err := s.toolbox.SearchGarments(ctx, filters, nameForSearch(namedGarment), 8)
		if err == nil && len(items) > 0 {
			summary := garmentListSummary(items)
			evidence = append(evidence, evidenceItem{
				Source: Source{
					Type: "馆藏检索", Title: "馆藏检索 · " + filterTitle(filters, namedGarment),
					Publisher: "非遗华裳馆藏", URL: "/explore", Snippet: summary,
				},
				Content: summary,
			})
			toolCalls++
			if err := s.emitTool(emit, "search_garments", fmt.Sprintf("检索到 %d 件馆藏", len(items)), toolStart); err != nil {
				return err
			}
		}
	}

	if contextGarment != nil {
		toolStart := time.Now()
		annotations, err := s.toolbox.GetAnnotations(ctx, contextGarment.ID)
		if err == nil && len(annotations) > 0 {
			limit := len(annotations)
			if limit > 3 {
				limit = 3
			}
			for _, a := range annotations[:limit] {
				evidence = append(evidence, evidenceItem{
					Source: Source{
						Type: "3D 讲解", Title: "3D 讲解 · " + a.Title,
						Publisher: "非遗华裳 3D 讲解",
						URL:       fmt.Sprintf("/detail/%d?annotation=%d", a.GarmentID, a.ID),
						Snippet:   firstSentence(a.Content, 160),
						GarmentID: a.GarmentID, AnnotationID: a.ID,
					},
					Content: a.Content,
				})
			}
			toolCalls++
			if err := s.emitTool(emit, "get_annotations", fmt.Sprintf("找到 %d 个 3D 热点", len(annotations)), toolStart); err != nil {
				return err
			}
		}
	}

	if contextGarment != nil && containsAny(message, []string{"相似", "类似", "推荐", "还有什么"}) {
		toolStart := time.Now()
		similar, err := s.toolbox.Similar(ctx, contextGarment.ID)
		if err == nil && len(similar) > 0 {
			summary := "相似馆藏：" + garmentListSummary(similar)
			evidence = append(evidence, evidenceItem{
				Source: Source{
					Type: "馆藏推荐", Title: "相似馆藏推荐",
					Publisher: "非遗华裳馆藏", URL: fmt.Sprintf("/detail/%d", contextGarment.ID), Snippet: summary,
				},
				Content: summary,
			})
			toolCalls++
			if err := s.emitTool(emit, "find_similar", fmt.Sprintf("找到 %d 件相似馆藏", len(similar)), toolStart); err != nil {
				return err
			}
		}
	}

	if queryGarment != nil && namedGarment != nil {
		// When the question names another garment, include its annotations too.
		annotations, err := s.toolbox.GetAnnotations(ctx, queryGarment.ID)
		if err == nil {
			limit := len(annotations)
			if limit > 2 {
				limit = 2
			}
			for _, a := range annotations[:limit] {
				evidence = append(evidence, evidenceItem{
					Source: Source{
						Type: "3D 讲解", Title: "3D 讲解 · " + a.Title,
						Publisher: "非遗华裳 3D 讲解",
						URL:       fmt.Sprintf("/detail/%d?annotation=%d", a.GarmentID, a.ID),
						Snippet:   firstSentence(a.Content, 160),
						GarmentID: a.GarmentID, AnnotationID: a.ID,
					},
					Content: a.Content,
				})
			}
			toolCalls++
		}
	}

	query := message
	if queryGarment != nil {
		query += " " + queryGarment.Name + " " + queryGarment.Ethnic + " " + queryGarment.Craft
	}
	if !filters.Empty() {
		query += " " + filters.Summary()
	}
	if statsIntent == "" {
		for _, chunk := range s.retriever.Search(ctx, query, 8) {
			meta := chunkMeta(chunk)
			typ := "资料"
			switch meta.SourceType {
			case "garment":
				typ = "馆藏"
			case "annotation":
				typ = "3D 讲解"
			case "article":
				typ = "资料"
			}
			evidence = append(evidence, evidenceItem{
				Source: Source{
					Type: typ, Title: chunk.Title, Publisher: chunk.Publisher,
					URL: chunk.SourceURL, Snippet: firstSentence(chunk.Content, 180),
					GarmentID: meta.GarmentID, AnnotationID: meta.AnnotationID,
				},
				Content: chunk.Content,
				Score:   chunk.Score,
			})
		}
	}

	evidence = dedupeEvidence(evidence, 8)
	sources := make([]Source, 0, len(evidence))
	for i := range evidence {
		evidence[i].Source.ID = fmt.Sprintf("S%d", i+1)
		sources = append(sources, evidence[i].Source)
	}
	if err := emit("sources", sources); err != nil {
		return err
	}

	if len(evidence) == 0 {
		answer := "馆藏资料中未找到直接依据。你可以换一个更具体的问法，例如指定服饰名称、民族、省份或工艺，我再从馆藏档案里帮你查。"
		if err := streamText(emit, answer); err != nil {
			return err
		}
		_ = emit("done", map[string]any{
			"conversation_id": conversationID, "message_id": messageID, "content": answer,
			"usage": Usage{}, "latency_ms": time.Since(start).Milliseconds(), "sources_used": []string{},
		})
		return nil
	}

	annotationForPrompt := ""
	if contextAnnotation != nil {
		annotationForPrompt = contextAnnotation.Title
	}
	userPrompt := buildUserPrompt(message, req.Context, evidence, garmentName(contextGarment), annotationForPrompt)
	messages := buildMessages(systemPrompt(), userPrompt, history)

	var (
		answer string
		usage  Usage
		err    error
	)
	if s.cfg.AIMockMode {
		answer = mockAnswer(evidence, message)
		if err := streamText(emit, answer); err != nil {
			return err
		}
	} else {
		answer, usage, err = s.provider.ChatStream(ctx, messages, func(delta string) error {
			return emit("delta", map[string]string{"text": delta})
		})
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("[AI][Chat] provider error: %v", err)
			return &ChatError{Code: 50201, Message: "AI 模型暂时不可用，请稍后重试", Retryable: true}
		}
	}

	finalText, usedSources := sanitizeCitations(answer, sources)
	finalText, highlightID := stripHighlight(finalText)
	if usage.PromptTokens == 0 {
		usage.PromptTokens = estimateTokens(userPrompt)
	}
	if usage.CompletionTokens == 0 {
		usage.CompletionTokens = estimateTokens(finalText)
	}
	usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	action := s.resolveAction(ctx, highlightID, contextAnnotation, evidence, finalText)
	if action != nil {
		if err := emit("action", action); err != nil {
			return err
		}
	}
	if err := emit("done", map[string]any{
		"conversation_id": conversationID, "message_id": messageID, "content": finalText,
		"usage": usage, "latency_ms": time.Since(start).Milliseconds(),
		"sources_used": usedSources,
	}); err != nil {
		return err
	}

	_ = s.usage.Record(context.Background(), UsageRecord{
		SessionHash: sessionHash, Endpoint: "chat", Model: s.cfg.AIChatModel,
		PromptTokens: usage.PromptTokens, CompletionTokens: usage.CompletionTokens,
		LatencyMS: time.Since(start).Milliseconds(), ToolCalls: toolCalls, Status: "ok",
	})
	return nil
}

func (s *Service) emitTool(emit EmitFunc, name, summary string, start time.Time) error {
	return emit("tool", ToolEvent{
		Name: name, Status: "done", Summary: summary,
		Elapsed: time.Since(start).Milliseconds(),
	})
}

func (s *Service) resolveAction(ctx context.Context, highlightID int64, contextAnnotation *model.Annotation, evidence []evidenceItem, answer string) *Action {
	if highlightID > 0 {
		if a, err := s.toolbox.annotationStore.GetByID(highlightID); err == nil {
			for _, item := range evidence {
				if item.Source.AnnotationID == a.ID {
					return &Action{Type: "highlight_annotation", GarmentID: a.GarmentID, AnnotationID: a.ID, Label: "在 3D 中查看「" + a.Title + "」"}
				}
			}
		}
	}
	if contextAnnotation != nil {
		return &Action{Type: "highlight_annotation", GarmentID: contextAnnotation.GarmentID, AnnotationID: contextAnnotation.ID, Label: "在 3D 中查看「" + contextAnnotation.Title + "」"}
	}
	for _, item := range evidence {
		if item.Source.AnnotationID > 0 && item.Source.Title != "" && strings.Contains(answer, strings.TrimPrefix(item.Source.Title, "3D 讲解 · ")) {
			if a, err := s.toolbox.annotationStore.GetByID(item.Source.AnnotationID); err == nil {
				return &Action{Type: "highlight_annotation", GarmentID: a.GarmentID, AnnotationID: a.ID, Label: "在 3D 中查看「" + a.Title + "」"}
			}
		}
	}
	return nil
}

func (s *Service) SaveFeedback(ctx context.Context, messageID string, rating int, reason string) error {
	if rating != 1 && rating != -1 {
		return &ChatError{Code: 40002, Message: "rating 只能是 1 或 -1", Retryable: false}
	}
	if strings.TrimSpace(messageID) == "" {
		return &ChatError{Code: 40001, Message: "缺少 message_id", Retryable: false}
	}
	return s.usage.SaveFeedback(ctx, messageID, rating, truncateRunes(reason, 300), s.cfg.AIChatModel)
}

func (s *Service) Reindex(ctx context.Context) (IngestStats, error) {
	stats, err := s.ingestor.Ingest(ctx, true, s.embedder)
	if err != nil {
		return stats, err
	}
	return stats, s.retriever.Refresh(ctx)
}

func dedupeEvidence(items []evidenceItem, limit int) []evidenceItem {
	seen := make(map[string]int)
	out := make([]evidenceItem, 0, len(items))
	for _, item := range items {
		key := item.Source.Type + "|" + fmt.Sprint(item.Source.GarmentID) + "|" + fmt.Sprint(item.Source.AnnotationID) + "|" + item.Source.Title
		if idx, ok := seen[key]; ok {
			if len([]rune(item.Content)) > len([]rune(out[idx].Content)) {
				out[idx] = item
			}
			continue
		}
		seen[key] = len(out)
		out = append(out, item)
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func sanitizeHistory(history []ChatTurn, max int) []ChatTurn {
	if max <= 0 {
		max = 8
	}
	if len(history) > max {
		history = history[len(history)-max:]
	}
	out := make([]ChatTurn, 0, len(history))
	total := 0
	for i := len(history) - 1; i >= 0; i-- {
		turn := history[i]
		role := strings.ToLower(strings.TrimSpace(turn.Role))
		content := strings.TrimSpace(turn.Content)
		if (role != "user" && role != "assistant") || content == "" {
			continue
		}
		content = truncateRunes(content, 1200)
		if total+len([]rune(content)) > 5000 {
			break
		}
		total += len([]rune(content))
		out = append([]ChatTurn{{Role: role, Content: content}}, out...)
	}
	return out
}

func streamText(emit EmitFunc, text string) error {
	runes := []rune(text)
	for i := 0; i < len(runes); i += 24 {
		end := i + 24
		if end > len(runes) {
			end = len(runes)
		}
		if err := emit("delta", map[string]string{"text": string(runes[i:end])}); err != nil {
			return err
		}
	}
	return nil
}

func statsDimension(q string) string {
	if !containsAny(q, []string{"统计", "分布", "最多", "数量", "排名", "多少个", "哪几个", "哪个省", "哪个民族", "哪些民族", "哪些工艺"}) {
		return ""
	}
	switch {
	case containsAny(q, []string{"省", "地区", "地域", "分布"}):
		return "province"
	case containsAny(q, []string{"民族"}):
		return "ethnic"
	case containsAny(q, []string{"工艺", "技艺"}):
		return "craft"
	case containsAny(q, []string{"朝代", "年代"}):
		return "dynasty"
	default:
		return ""
	}
}

func statsSummary(dim string, items []StatsItem) string {
	var b strings.Builder
	fmt.Fprintf(&b, "馆藏按%s统计（前 %d 组）：", dimensionLabel(dim), len(items))
	for i, item := range items {
		if i > 0 {
			b.WriteString("；")
		}
		fmt.Fprintf(&b, "%s %d 件", item.Label, item.Count)
	}
	b.WriteString("。")
	return b.String()
}

func dimensionLabel(dim string) string {
	switch dim {
	case "province":
		return "省份"
	case "ethnic":
		return "民族"
	case "craft":
		return "工艺"
	case "dynasty":
		return "朝代"
	default:
		return dim
	}
}

func garmentListSummary(items []GarmentBrief) string {
	var b strings.Builder
	for i, item := range items {
		if i > 0 {
			b.WriteString("；")
		}
		fmt.Fprintf(&b, "%s（%s，%s，%s）", item.Name, emptyAs(item.Ethnic, "民族未标注"), emptyAs(item.Craft, "工艺未标注"), emptyAs(item.Province, "地域未标注"))
	}
	return b.String()
}

func filterTitle(filters FilterSet, named *GarmentBrief) string {
	if named != nil {
		return named.Name
	}
	if s := filters.Summary(); s != "" {
		return s
	}
	return "全部馆藏"
}

func nameForSearch(named *GarmentBrief) string {
	if named == nil {
		return ""
	}
	return named.Name
}

func garmentName(g *model.Garment) string {
	if g == nil {
		return ""
	}
	return g.Name
}

func containsAny(s string, words []string) bool {
	for _, word := range words {
		if strings.Contains(s, word) {
			return true
		}
	}
	return false
}

func estimateTokens(s string) int {
	n := len([]rune(s))
	if n == 0 {
		return 0
	}
	return n/2 + 1
}

func randomID(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s%d", prefix, time.Now().UnixNano())
	}
	return prefix + hex.EncodeToString(buf)
}

func hashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:8])
}

func statusForError(err error) string {
	var chatErr *ChatError
	if errors.As(err, &chatErr) {
		return fmt.Sprintf("error_%d", chatErr.Code)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	return "error"
}
