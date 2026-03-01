package worker

import "strings"

// ChunkTranscript splits a transcript into chunks of approximately maxTokens size
// with overlap tokens of overlap between consecutive chunks.
// Token estimation: split by whitespace, ~1 word ≈ 1 token.
func ChunkTranscript(text string, maxTokens, overlap int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	if len(words) <= maxTokens {
		return []string{text}
	}

	var chunks []string
	start := 0
	for start < len(words) {
		end := start + maxTokens
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, strings.Join(words[start:end], " "))
		if end == len(words) {
			break
		}
		start = end - overlap
		if start < 0 {
			start = 0
		}
	}
	return chunks
}
