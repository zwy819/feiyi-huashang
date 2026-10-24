package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"nonlegacy-garment-3d/server/internal/model"
)

// Embedder is implemented by the OpenAI-compatible provider. A nil embedder
// means the corpus is still fully usable through BM25; vectors are additive.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	EmbeddingModel() string
}

type IngestStats struct {
	Documents int `json:"documents"`
	Chunks    int `json:"chunks"`
	Embedded  int `json:"embedded"`
	Skipped   int `json:"skipped"`
}

type Ingestor struct {
	store           *KnowledgeStore
	garmentStore    *model.GarmentStore
	annotationStore *model.AnnotationStore
	articlesDir     string
}

func NewIngestor(
	store *KnowledgeStore,
	garmentStore *model.GarmentStore,
	annotationStore *model.AnnotationStore,
	articlesDir string,
) *Ingestor {
	return &Ingestor{
		store:           store,
		garmentStore:    garmentStore,
		annotationStore: annotationStore,
		articlesDir:     articlesDir,
	}
}

type pendingDocument struct {
	doc      KnowledgeDocument
	metadata map[string]any
	chunks   []KnowledgeChunk
}

type pendingChunk struct {
	docIndex   int
	chunkIndex int
	content    string
	metadata   map[string]any
}

// Ingest builds or refreshes the retrieval corpus. It is safe to call on every
// startup: unchanged documents are skipped by checksum, and Force rebuilds all.
func (ing *Ingestor) Ingest(ctx context.Context, force bool, embedder Embedder) (IngestStats, error) {
	stats := IngestStats{}
	if force {
		if err := ing.store.ClearAll(ctx); err != nil {
			return stats, fmt.Errorf("clear knowledge base: %w", err)
		}
	}

	docs, err := ing.collectDocuments(ctx)
	if err != nil {
		return stats, err
	}

	pending := make([]pendingDocument, 0, len(docs))
	for _, doc := range docs {
		id, checksum, found, err := ing.store.GetDocumentMeta(ctx, doc.DocKey)
		if err != nil {
			return stats, err
		}
		if found && checksum == doc.Checksum {
			if n, err := ing.store.ChunkCount(ctx, id); err == nil && n > 0 {
				stats.Skipped++
				continue
			}
		}

		docID, err := ing.store.UpsertDocument(ctx, doc)
		if err != nil {
			return stats, fmt.Errorf("upsert %s: %w", doc.DocKey, err)
		}
		chunks := ChunkText(doc.Content, defaultChunkSize, defaultChunkOverlap)
		if len(chunks) == 0 {
			log.Printf("[AI][Ingest] 跳过空文档 %s", doc.DocKey)
			continue
		}
		item := pendingDocument{doc: doc, metadata: docMetadata(doc), chunks: make([]KnowledgeChunk, 0, len(chunks))}
		for i, text := range chunks {
			item.chunks = append(item.chunks, KnowledgeChunk{
				DocumentID:   docID,
				ChunkIndex:   i,
				Content:      text,
				TokenCount:   runeCount(text),
				MetadataJSON: mustJSON(item.metadata),
			})
		}
		pending = append(pending, item)
		stats.Documents++
		stats.Chunks += len(chunks)
	}

	if len(pending) == 0 {
		return stats, nil
	}

	if embedder != nil {
		flat := make([]pendingChunk, 0, stats.Chunks)
		for di := range pending {
			for ci := range pending[di].chunks {
				flat = append(flat, pendingChunk{
					docIndex:   di,
					chunkIndex: ci,
					content:    pending[di].chunks[ci].Content,
					metadata:   pending[di].metadata,
				})
			}
		}
		vectors := make([][]float32, len(flat))
		failed := false
		for start := 0; start < len(flat); start += 10 {
			end := start + 10
			if end > len(flat) {
				end = len(flat)
			}
			texts := make([]string, 0, end-start)
			for _, item := range flat[start:end] {
				texts = append(texts, truncateRunes(item.content, 1800))
			}
			batch, err := embedder.Embed(ctx, texts)
			if err != nil {
				log.Printf("[AI][Ingest] 向量化失败，继续使用 BM25: %v", err)
				failed = true
				break
			}
			if len(batch) != len(texts) {
				log.Printf("[AI][Ingest] 向量数量不匹配: got=%d want=%d", len(batch), len(texts))
				failed = true
				break
			}
			copy(vectors[start:end], batch)
		}
		if !failed {
			for i, item := range flat {
				pending[item.docIndex].chunks[item.chunkIndex].Embedding = vectors[i]
				pending[item.docIndex].chunks[item.chunkIndex].EmbeddingModel = embedder.EmbeddingModel()
			}
			stats.Embedded = len(flat)
		}
	}

	for _, item := range pending {
		if err := ing.store.ReplaceChunks(ctx, item.chunks[0].DocumentID, item.chunks); err != nil {
			return stats, fmt.Errorf("replace chunks for %s: %w", item.doc.DocKey, err)
		}
	}

	return stats, nil
}

func (ing *Ingestor) collectDocuments(ctx context.Context) ([]KnowledgeDocument, error) {
	articles, err := LoadArticles(ing.articlesDir)
	if err != nil {
		return nil, fmt.Errorf("load articles: %w", err)
	}

	docs := make([]KnowledgeDocument, 0, len(articles)+180)
	for _, article := range articles {
		content := article.Content
		if article.Tags != "" {
			content = "关键词：" + article.Tags + "\n\n" + content
		}
		docs = append(docs, KnowledgeDocument{
			DocKey:     "article:" + article.Slug,
			Title:      article.Title,
			SourceType: "article",
			SourceURL:  article.SourceURL,
			Publisher:  article.Publisher,
			License:    article.License,
			Content:    content,
			Checksum:   checksum(article.Title + "\n" + article.SourceURL + "\n" + content),
			Status:     "active",
		})
	}

	garments, _, err := ing.garmentStore.List(model.GarmentQueryParams{Page: 1, Limit: 1000})
	if err != nil {
		return nil, fmt.Errorf("list garments: %w", err)
	}
	for i := range garments {
		g := &garments[i]
		content := garmentDocument(*g)
		docs = append(docs, KnowledgeDocument{
			DocKey:     fmt.Sprintf("garment:%d", g.ID),
			Title:      "馆藏档案 · " + g.Name,
			SourceType: "garment",
			SourceURL:  fmt.Sprintf("/detail/%d", g.ID),
			Publisher:  "非遗华裳馆藏",
			License:    "项目自建馆藏说明，仅用于本作品展示",
			Content:    content,
			Checksum:   checksum(content),
			Status:     "active",
			GarmentID:  g.ID,
			Ethnic:     g.Ethnic,
			Province:   g.Province,
			Craft:      g.Craft,
		})
	}

	annotations, err := ing.annotationStore.List(0)
	if err != nil {
		return nil, fmt.Errorf("list annotations: %w", err)
	}
	for i := range annotations {
		a := &annotations[i]
		content := fmt.Sprintf("服饰：%s\n讲解点：%s\n%s", a.GarmentName, a.Title, a.Content)
		docs = append(docs, KnowledgeDocument{
			DocKey:       fmt.Sprintf("annotation:%d", a.ID),
			Title:        "3D 讲解 · " + a.Title,
			SourceType:   "annotation",
			SourceURL:    fmt.Sprintf("/detail/%d?annotation=%d", a.GarmentID, a.ID),
			Publisher:    "非遗华裳 3D 讲解",
			License:      "项目自建 3D 标注，仅用于本作品展示",
			Content:      content,
			Checksum:     checksum(content),
			Status:       "active",
			GarmentID:    a.GarmentID,
			AnnotationID: a.ID,
		})
	}

	return docs, nil
}

func garmentDocument(g model.Garment) string {
	var b strings.Builder
	fmt.Fprintf(&b, "名称：%s\n", g.Name)
	if g.Dynasty != "" {
		fmt.Fprintf(&b, "朝代：%s\n", g.Dynasty)
	}
	if g.Ethnic != "" {
		fmt.Fprintf(&b, "民族：%s\n", g.Ethnic)
	}
	if g.Craft != "" {
		fmt.Fprintf(&b, "工艺：%s\n", g.Craft)
	}
	if g.Material != "" {
		fmt.Fprintf(&b, "材质：%s\n", g.Material)
	}
	if g.Pattern != "" {
		fmt.Fprintf(&b, "纹样：%s\n", g.Pattern)
	}
	if g.Province != "" {
		fmt.Fprintf(&b, "地域：%s\n", g.Province)
	}
	if g.CulturalMeaning != "" {
		fmt.Fprintf(&b, "文化寓意：%s\n", g.CulturalMeaning)
	}
	if g.Description != "" {
		fmt.Fprintf(&b, "馆藏说明：%s\n", g.Description)
	}
	return b.String()
}

func docMetadata(doc KnowledgeDocument) map[string]any {
	meta := map[string]any{
		"doc_key":     doc.DocKey,
		"source_type": doc.SourceType,
	}
	if doc.GarmentID > 0 {
		meta["garment_id"] = doc.GarmentID
	}
	if doc.AnnotationID > 0 {
		meta["annotation_id"] = doc.AnnotationID
	}
	if doc.Ethnic != "" {
		meta["ethnic"] = doc.Ethnic
	}
	if doc.Province != "" {
		meta["province"] = doc.Province
	}
	if doc.Craft != "" {
		meta["craft"] = doc.Craft
	}
	return meta
}

func checksum(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func mustJSON(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}
