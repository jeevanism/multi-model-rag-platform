package rag

import "strings"

func BuildGroundedPrompt(question string, chunks []RetrievedChunk) string {
	if len(chunks) == 0 {
		return question
	}

	contextLines := make([]string, 0, len(chunks))
	citations := FormatCitations(chunks)
	for i, chunk := range chunks {
		contextLines = append(contextLines, citations[i]+" "+chunk.Content)
	}

	context := strings.Join(contextLines, "\n")
	return "Answer the question using only the provided context. " +
		"Include citations using the provided source tags.\n\n" +
		"Context:\n" + context + "\n\nQuestion:\n" + question
}
