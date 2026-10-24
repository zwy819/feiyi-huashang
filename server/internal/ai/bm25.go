package ai

import (
	"math"
	"sort"
	"strings"
)

type bm25Doc struct {
	chunk RetrievedChunk
	tf    map[string]int
	len   int
}

// BM25Index is a deliberately small in-process index: 200–300 chunks do not
// justify an external search engine, and rebuilding it takes milliseconds.
type BM25Index struct {
	docs  []bm25Doc
	df    map[string]int
	avgdl float64
}

func NewBM25Index(chunks []RetrievedChunk) *BM25Index {
	idx := &BM25Index{docs: make([]bm25Doc, 0, len(chunks)), df: make(map[string]int)}
	totalLen := 0
	for _, chunk := range chunks {
		tokens := Tokenize(chunk.Title + "\n" + chunk.Content)
		doc := bm25Doc{chunk: chunk, tf: termFrequency(tokens), len: len(tokens)}
		for term := range doc.tf {
			idx.df[term]++
		}
		totalLen += doc.len
		idx.docs = append(idx.docs, doc)
	}
	if len(idx.docs) > 0 {
		idx.avgdl = float64(totalLen) / float64(len(idx.docs))
	}
	return idx
}

func (idx *BM25Index) Size() int {
	return len(idx.docs)
}

// Score returns all documents sorted by BM25 relevance. The caller applies the
// cutoff and fusion; keeping the full list makes RRF simple.
func (idx *BM25Index) Score(query string) []RetrievedChunk {
	terms := uniqueTerms(query)
	if len(terms) == 0 || len(idx.docs) == 0 {
		return nil
	}

	const k1, b = 1.2, 0.75
	n := float64(len(idx.docs))
	lowerQuery := strings.ToLower(strings.TrimSpace(query))
	scores := make([]RetrievedChunk, 0, len(idx.docs))

	for _, doc := range idx.docs {
		var score float64
		for _, term := range terms {
			tf := float64(doc.tf[term])
			if tf == 0 {
				continue
			}
			df := float64(idx.df[term])
			idf := math.Log(1 + (n-df+0.5)/(df+0.5))
			denom := tf + k1*(1-b+b*float64(doc.len)/math.Max(idx.avgdl, 1))
			score += idf * (tf * (k1 + 1)) / denom
		}
		if score == 0 {
			continue
		}
		if len([]rune(lowerQuery)) >= 2 {
			if strings.Contains(strings.ToLower(doc.chunk.Content), lowerQuery) {
				score += 3.5
			}
			if strings.Contains(strings.ToLower(doc.chunk.Title), lowerQuery) {
				score += 2.0
			}
		}
		out := doc.chunk
		out.BM25 = score
		out.Score = score
		scores = append(scores, out)
	}

	sort.SliceStable(scores, func(i, j int) bool {
		if scores[i].BM25 == scores[j].BM25 {
			return scores[i].ID < scores[j].ID
		}
		return scores[i].BM25 > scores[j].BM25
	})
	return scores
}

// RRF fuses keyword and vector rankings without having to calibrate two
// incomparable score distributions.
func RRF(keyword, vector []RetrievedChunk, k int) []RetrievedChunk {
	if k <= 0 {
		k = 60
	}
	type fused struct {
		chunk RetrievedChunk
		score float64
	}
	byID := make(map[int64]*fused)
	add := func(list []RetrievedChunk) {
		for rank, item := range list {
			entry, ok := byID[item.ID]
			if !ok {
				entry = &fused{chunk: item}
				byID[item.ID] = entry
			}
			entry.score += 1 / float64(k+rank+1)
			if item.BM25 > entry.chunk.BM25 {
				entry.chunk.BM25 = item.BM25
			}
			if item.Vector > entry.chunk.Vector {
				entry.chunk.Vector = item.Vector
			}
		}
	}
	add(keyword)
	add(vector)

	out := make([]RetrievedChunk, 0, len(byID))
	for _, entry := range byID {
		entry.chunk.Score = entry.score
		out = append(out, entry.chunk)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].BM25 > out[j].BM25
		}
		return out[i].Score > out[j].Score
	})
	return out
}

func cosine(a, b []float32) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		x, y := float64(a[i]), float64(b[i])
		dot += x * y
		na += x * x
		nb += y * y
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}
