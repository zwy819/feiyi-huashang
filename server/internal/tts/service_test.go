package tts

import (
	"strings"
	"testing"

	"nonlegacy-garment-3d/server/internal/model"
)

func TestGarmentNarration(t *testing.T) {
	item := &model.Garment{
		Name:            "苗族百褶裙",
		Dynasty:         "清",
		Ethnic:          "苗族",
		Province:        "贵州",
		Craft:           "蜡染+刺绣",
		Material:        "棉麻",
		Pattern:         "几何纹+花卉纹",
		CulturalMeaning: "百褶代表山川河流的绵延不绝",
	}
	text := garmentNarration(item)
	for _, want := range []string{"苗族百褶裙", "清代", "蜡染+刺绣", "百褶代表"} {
		if !strings.Contains(text, want) {
			t.Fatalf("narration missing %q: %s", want, text)
		}
	}
	if strings.Contains(text, "清代代") {
		t.Fatalf("dynasty duplicated: %s", text)
	}
}

func TestNormalizeNarration(t *testing.T) {
	if got := normalizeNarration("  你好   世界  "); got != "你好 世界" {
		t.Fatalf("normalize = %q", got)
	}
	long := strings.Repeat("服", 620)
	got := normalizeNarration(long)
	if len([]rune(got)) != 601 {
		t.Fatalf("truncated length = %d, want 601 with punctuation", len([]rune(got)))
	}
	if !strings.HasSuffix(got, "。") {
		t.Fatalf("truncated text should end with punctuation: %q", got[len(got)-10:])
	}
}

func TestCacheKey(t *testing.T) {
	a := cacheKey("qwen3-tts-flash", "Cherry", "你好")
	b := cacheKey("qwen3-tts-flash", "Cherry", "你好")
	c := cacheKey("qwen3-tts-flash", "Serena", "你好")
	if a != b {
		t.Fatal("cache key should be deterministic")
	}
	if a == c {
		t.Fatal("cache key should include voice")
	}
	if len(a) != 32 {
		t.Fatalf("cache key length = %d", len(a))
	}
}
