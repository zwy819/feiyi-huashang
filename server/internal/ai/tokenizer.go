package ai

import (
	"strings"
	"unicode"
)

// Tokenize produces a mix of Han unigrams/bigrams and lower-cased Latin or
// numeric words. Character bigrams avoid a Chinese word segmenter while still
// preserving enough phrase information for BM25.
func Tokenize(s string) []string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return nil
	}

	tokens := make([]string, 0, len([]rune(s)))
	var latin []rune
	var prevHan rune

	flushLatin := func() {
		if len(latin) > 0 {
			tokens = append(tokens, string(latin))
			latin = latin[:0]
		}
	}

	for _, r := range s {
		switch {
		case unicode.Is(unicode.Han, r):
			flushLatin()
			tokens = append(tokens, string(r))
			if prevHan != 0 {
				tokens = append(tokens, string([]rune{prevHan, r}))
			}
			prevHan = r
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			prevHan = 0
			latin = append(latin, r)
		default:
			flushLatin()
			prevHan = 0
		}
	}
	flushLatin()

	return tokens
}

func termFrequency(tokens []string) map[string]int {
	tf := make(map[string]int, len(tokens))
	for _, token := range tokens {
		tf[token]++
	}
	return tf
}

func uniqueTerms(s string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, token := range Tokenize(s) {
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, token)
	}
	return out
}
