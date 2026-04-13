package rag

import "fmt"

func FormatCitations(chunks []RetrievedChunk) []string {
	citations := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		citations = append(citations, fmt.Sprintf("[source:%s#chunk=%d]", chunk.Title, chunk.ChunkIndex))
	}
	return citations
}
