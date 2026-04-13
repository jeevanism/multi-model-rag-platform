package rag

import (
	"errors"
	"strings"
)

func ChunkText(text string, maxChars int, overlapChars int) ([]string, error) {
	normalized := strings.Join(strings.Fields(text), " ")
	if normalized == "" {
		return []string{}, nil
	}
	if maxChars <= 0 {
		return nil, errors.New("max_chars must be > 0")
	}
	if overlapChars < 0 {
		return nil, errors.New("overlap_chars must be >= 0")
	}
	if overlapChars >= maxChars {
		return nil, errors.New("overlap_chars must be smaller than max_chars")
	}

	chunks := make([]string, 0)
	start := 0
	length := len(normalized)

	for start < length {
		end := min(start+maxChars, length)
		if end < length {
			splitAt := strings.LastIndex(normalized[start:end], " ")
			if splitAt > 0 {
				end = start + splitAt
			}
		}

		chunk := strings.TrimSpace(normalized[start:end])
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
		if end >= length {
			break
		}

		start = max(0, end-overlapChars)
		for start < length && normalized[start] == ' ' {
			start++
		}
	}

	return chunks, nil
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
