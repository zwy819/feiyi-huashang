package ai

import (
	"strings"
	"testing"
)

func TestTokenizeChineseAndLatin(t *testing.T) {
	tokens := Tokenize("苗绣 Miao embroidery 2026")
	joined := strings.Join(tokens, "|")
	for _, want := range []string{"苗", "苗绣", "miao", "embroidery", "2026"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("tokens %q missing %q", joined, want)
		}
	}
}

func TestChunkTextOverlapAndBoundaries(t *testing.T) {
	text := strings.Repeat("苗族百褶裙的裙褶象征迁徙记忆。", 60)
	chunks := ChunkText(text, 200, 40)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
	for _, chunk := range chunks {
		if n := runeCount(chunk); n > 220 {
			t.Fatalf("chunk too large: %d", n)
		}
		if !strings.HasSuffix(chunk, "。") {
			t.Fatalf("chunk did not end on a sentence boundary: %q", chunk[len(chunk)-10:])
		}
	}
}

func TestParseArticle(t *testing.T) {
	raw := `---
title: 苏绣
publisher: 中国非物质文化遗产网
source_url: https://www.ihchina.cn/art/detail/id/13978.html
license: 事实摘要
tags: 苏绣, 江苏
---
苏绣是苏州地区代表性的刺绣工艺，以针法细密、色彩雅致著称。`
	article, err := ParseArticle(raw)
	if err != nil {
		t.Fatal(err)
	}
	if article.Title != "苏绣" || article.SourceURL == "" || article.Tags == "" {
		t.Fatalf("unexpected article: %+v", article)
	}
	if !strings.Contains(article.Content, "针法细密") {
		t.Fatalf("content not parsed: %q", article.Content)
	}
}

func TestRRFKeepsBothRetrievers(t *testing.T) {
	keyword := []RetrievedChunk{{Chunk: Chunk{ID: 1}, BM25: 4}, {Chunk: Chunk{ID: 2}, BM25: 2}}
	vector := []RetrievedChunk{{Chunk: Chunk{ID: 2}, Vector: 0.9}, {Chunk: Chunk{ID: 3}, Vector: 0.8}}
	fused := RRF(keyword, vector, 60)
	if len(fused) != 3 {
		t.Fatalf("expected 3 fused chunks, got %d", len(fused))
	}
	if fused[0].ID != 2 {
		t.Fatalf("chunk 2 should rank first, got %d", fused[0].ID)
	}
}

func TestBundledArticles(t *testing.T) {
	articles, err := LoadArticles("../../knowledge/articles")
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) < 20 {
		t.Fatalf("expected at least 20 sourced articles, got %d", len(articles))
	}
	for _, article := range articles {
		if article.SourceURL == "" {
			t.Fatalf("%s has no source URL", article.Slug)
		}
		if runeCount(article.Content) < 200 {
			t.Fatalf("%s is too short: %d runes", article.Slug, runeCount(article.Content))
		}
	}
}

func TestSanitizeCitationsDropsInvalidHandles(t *testing.T) {
	sources := []Source{{ID: "S1"}, {ID: "S2"}}
	text, used := sanitizeCitations("蜡染用靛蓝染色[S1]，传说内容[S9]，补充[S2]。", sources)
	if strings.Contains(text, "S9") {
		t.Fatalf("invalid citation was not removed: %q", text)
	}
	if len(used) != 2 || used[0] != "S1" || used[1] != "S2" {
		t.Fatalf("unexpected used sources: %v", used)
	}
}

func TestStripHighlight(t *testing.T) {
	text, id := stripHighlight("去 3D 看细节。\n[[HIGHLIGHT:12]]")
	if id != 12 {
		t.Fatalf("expected highlight 12, got %d", id)
	}
	if strings.Contains(text, "HIGHLIGHT") {
		t.Fatalf("marker leaked into text: %q", text)
	}

	text, id = stripHighlight("看看细节。\n[[HIGHLIGHT:3D_讲解_百褶工艺]]")
	if id != 0 {
		t.Fatalf("non-numeric highlight should not resolve, got %d", id)
	}
	if strings.Contains(text, "HIGHLIGHT") {
		t.Fatalf("non-numeric marker leaked into text: %q", text)
	}
}

func TestDynastyFilterAvoidsSubstringNoise(t *testing.T) {
	if matchesDimension("画面很清晰", "dynasty", "清") {
		t.Fatal("single-character dynasty must not match inside an unrelated word")
	}
	if !matchesDimension("清代服饰有什么特点", "dynasty", "清") {
		t.Fatal("清代 should match the 清 dynasty")
	}
	if !matchesDimension("苗族百褶裙", "ethnic", "苗族") {
		t.Fatal("ethnic substring should match")
	}
}
