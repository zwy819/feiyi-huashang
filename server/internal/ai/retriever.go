package ai

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"strings"
	"sync"
)

type ChunkMetadata struct {
	DocKey       string `json:"doc_key"`
	SourceType   string `json:"source_type"`
	GarmentID    int64  `json:"garment_id"`
	AnnotationID int64  `json:"annotation_id"`
	Ethnic       string `json:"ethnic"`
	Province     string `json:"province"`
	Craft        string `json:"craft"`
}

type Retriever struct {
	store    *KnowledgeStore
	embedder Embedder

	mu     sync.RWMutex
	chunks []RetrievedChunk
	index  *BM25Index
}

func NewRetriever(store *KnowledgeStore, embedder Embedder) *Retriever {
	return &Retriever{store: store, embedder: embedder}
}

func (r *Retriever) Refresh(ctx context.Context) error {
	chunks, err := r.store.LoadChunks(ctx)
	if err != nil {
		return err
	}
	idx := NewBM25Index(chunks)
	r.mu.Lock()
	r.chunks = chunks
	r.index = idx
	r.mu.Unlock()
	log.Printf("[AI][RAG] 已加载 %d 个知识块", len(chunks))
	return nil
}

func (r *Retriever) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.chunks)
}

func (r *Retriever) Search(ctx context.Context, query string, topK int) []RetrievedChunk {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	if topK <= 0 {
		topK = 6
	}

	r.mu.RLock()
	idx := r.index
	chunks := r.chunks
	r.mu.RUnlock()
	if idx == nil || len(chunks) == 0 {
		return nil
	}

	keyword := idx.Score(query)
	if len(keyword) > 30 {
		keyword = keyword[:30]
	}

	var vector []RetrievedChunk
	if r.embedder != nil {
		vecs, err := r.embedder.Embed(ctx, []string{truncateRunes(query, 600)})
		if err != nil {
			log.Printf("[AI][RAG] 查询向量化失败，仅使用 BM25: %v", err)
		} else if len(vecs) == 1 && len(vecs[0]) > 0 {
			for _, chunk := range chunks {
				if len(chunk.Embedding) == 0 {
					continue
				}
				score := cosine(vecs[0], chunk.Embedding)
				if score <= 0 {
					continue
				}
				out := chunk
				out.Vector = score
				out.Score = score
				vector = append(vector, out)
			}
			sort.SliceStable(vector, func(i, j int) bool {
				if vector[i].Vector == vector[j].Vector {
					return vector[i].ID < vector[j].ID
				}
				return vector[i].Vector > vector[j].Vector
			})
			if len(vector) > 30 {
				vector = vector[:30]
			}
		}
	}

	fused := RRF(keyword, vector, 60)
	if len(fused) > topK {
		fused = fused[:topK]
	}
	return fused
}

func chunkMeta(chunk RetrievedChunk) ChunkMetadata {
	var meta ChunkMetadata
	if chunk.MetadataJSON != "" {
		_ = json.Unmarshal([]byte(chunk.MetadataJSON), &meta)
	}
	if meta.SourceType == "" {
		meta.SourceType = chunk.SourceType
	}
	return meta
}
