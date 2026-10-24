package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
)

// KnowledgeStore owns the AI-only tables. The rest of the application keeps
// using the existing stores; the AI module only reads garment/annotation data.
type KnowledgeStore struct {
	db *sql.DB
}

func NewKnowledgeStore(db *sql.DB) *KnowledgeStore {
	return &KnowledgeStore{db: db}
}

func (s *KnowledgeStore) CountDocuments(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM knowledge_documents WHERE status='active'`).Scan(&n)
	return n, err
}

func (s *KnowledgeStore) CountChunks(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM knowledge_chunks`).Scan(&n)
	return n, err
}

// UpsertDocument inserts or updates a document by doc_key and returns its id.
func (s *KnowledgeStore) UpsertDocument(ctx context.Context, doc KnowledgeDocument) (int64, error) {
	if doc.DocKey == "" || doc.Title == "" {
		return 0, errors.New("knowledge document requires doc_key and title")
	}
	if doc.Status == "" {
		doc.Status = "active"
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO knowledge_documents
			(doc_key, title, source_type, source_url, publisher, license, content, checksum, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(doc_key) DO UPDATE SET
			title=excluded.title,
			source_type=excluded.source_type,
			source_url=excluded.source_url,
			publisher=excluded.publisher,
			license=excluded.license,
			content=excluded.content,
			checksum=excluded.checksum,
			status=excluded.status,
			updated_at=CURRENT_TIMESTAMP`,
		doc.DocKey, doc.Title, doc.SourceType, doc.SourceURL, doc.Publisher,
		doc.License, doc.Content, doc.Checksum, doc.Status,
	)
	if err != nil {
		return 0, err
	}

	var id int64
	err = s.db.QueryRowContext(ctx, `SELECT id FROM knowledge_documents WHERE doc_key = ?`, doc.DocKey).Scan(&id)
	return id, err
}

// ReplaceChunks replaces all chunks of a document once its content changed.
func (s *KnowledgeStore) ReplaceChunks(ctx context.Context, documentID int64, chunks []KnowledgeChunk) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM knowledge_chunks WHERE document_id = ?`, documentID); err != nil {
		return err
	}
	for _, chunk := range chunks {
		embedding := ""
		if len(chunk.Embedding) > 0 {
			raw, err := json.Marshal(chunk.Embedding)
			if err != nil {
				return err
			}
			embedding = string(raw)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO knowledge_chunks
				(document_id, chunk_index, content, token_count, embedding, embedding_model, metadata_json)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			documentID, chunk.ChunkIndex, chunk.Content, chunk.TokenCount,
			embedding, chunk.EmbeddingModel, chunk.MetadataJSON,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *KnowledgeStore) UpdateChunkEmbedding(ctx context.Context, chunkID int64, embedding []float32, model string) error {
	raw, err := json.Marshal(embedding)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE knowledge_chunks SET embedding=?, embedding_model=? WHERE id=?`,
		string(raw), model, chunkID)
	return err
}

// LoadChunks returns every active chunk with its document citation metadata.
func (s *KnowledgeStore) LoadChunks(ctx context.Context) ([]RetrievedChunk, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, c.document_id, c.chunk_index, c.content, c.token_count,
		       c.embedding, c.embedding_model, c.metadata_json,
		       d.title, d.source_type, d.source_url, d.publisher, d.license
		FROM knowledge_chunks c
		JOIN knowledge_documents d ON d.id = c.document_id
		WHERE d.status = 'active'
		ORDER BY c.document_id, c.chunk_index`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chunks := make([]RetrievedChunk, 0)
	for rows.Next() {
		var c RetrievedChunk
		var embedding string
		if err := rows.Scan(
			&c.ID, &c.DocumentID, &c.ChunkIndex, &c.Content, &c.TokenCount,
			&embedding, &c.EmbeddingModel, &c.MetadataJSON,
			&c.Title, &c.SourceType, &c.SourceURL, &c.Publisher, &c.License,
		); err != nil {
			return nil, err
		}
		if embedding != "" {
			_ = json.Unmarshal([]byte(embedding), &c.Embedding)
		}
		chunks = append(chunks, c)
	}
	return chunks, rows.Err()
}

func (s *KnowledgeStore) ClearAll(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM knowledge_documents`)
	if err == nil {
		_, err = s.db.ExecContext(ctx, `DELETE FROM knowledge_chunks`)
	}
	return err
}

// GetDocumentMeta returns the id and checksum of an existing document.
func (s *KnowledgeStore) GetDocumentMeta(ctx context.Context, docKey string) (id int64, checksum string, found bool, err error) {
	err = s.db.QueryRowContext(ctx, `SELECT id, checksum FROM knowledge_documents WHERE doc_key = ?`, docKey).Scan(&id, &checksum)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", false, nil
	}
	if err != nil {
		return 0, "", false, err
	}
	return id, checksum, true, nil
}

func (s *KnowledgeStore) ChunkCount(ctx context.Context, documentID int64) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM knowledge_chunks WHERE document_id = ?`, documentID).Scan(&n)
	return n, err
}
