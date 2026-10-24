package ai

import "strings"

const (
	defaultChunkSize    = 480
	defaultChunkOverlap = 80
)

// ChunkText splits Chinese prose into overlapping windows, preferring sentence
// and paragraph boundaries. Sizes are measured in runes, which is a close
// enough proxy for Chinese token count at this corpus size.
func ChunkText(text string, size, overlap int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if size <= 0 {
		size = defaultChunkSize
	}
	if overlap < 0 || overlap >= size {
		overlap = defaultChunkOverlap
	}

	runes := []rune(text)
	if len(runes) <= size {
		return []string{text}
	}

	var chunks []string
	for start := 0; start < len(runes); {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		if end < len(runes) {
			// Look back for a natural break, but never cut the chunk in half.
			for i := end; i > start+size/2; i-- {
				if isNaturalBreak(runes[i-1]) {
					end = i
					break
				}
			}
		}

		chunk := strings.TrimSpace(string(runes[start:end]))
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
		if end >= len(runes) {
			break
		}

		next := end - overlap
		if next <= start {
			next = end
		}
		start = next
	}

	return chunks
}

func isNaturalBreak(r rune) bool {
	switch r {
	case '。', '！', '？', '；', '\n', '!', '?', ';':
		return true
	default:
		return false
	}
}

func runeCount(s string) int {
	return len([]rune(s))
}
