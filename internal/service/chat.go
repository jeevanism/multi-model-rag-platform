package service

import (
	"errors"

	"multi-model-rag-platform/internal/llm"
	"multi-model-rag-platform/internal/rag"
)

var ErrRAGRepositoryRequired = errors.New("Database session is required when rag=true")

type ChatParams struct {
	Message  string
	Provider string
	Model    *string
	RAG      bool
	TopK     int
	Debug    bool
}

type ChatResult struct {
	Response        llm.Response
	Citations       []string
	RAGUsed         bool
	RetrievedChunks []rag.RetrievedChunk
}

type ChatRepository interface {
	RetrieveChunks(query string, topK int) ([]rag.RetrievedChunk, error)
}

type ChatService struct {
	router llm.Router
	repo   ChatRepository
}

func NewChatService(router llm.Router, repo ChatRepository) ChatService {
	return ChatService{router: router, repo: repo}
}

func (s ChatService) GenerateChatResponse(req ChatParams, forceStub bool) (ChatResult, error) {
	prompt := req.Message
	citations := []string{}
	retrievedChunks := []rag.RetrievedChunk(nil)
	ragUsed := false

	if req.RAG {
		if s.repo == nil {
			return ChatResult{}, ErrRAGRepositoryRequired
		}

		chunks, err := s.repo.RetrieveChunks(req.Message, req.TopK)
		if err != nil {
			return ChatResult{}, err
		}
		retrievedChunks = chunks
		citations = rag.FormatCitations(chunks)
		prompt = rag.BuildGroundedPrompt(req.Message, chunks)
		ragUsed = true
	}

	provider, err := s.router.GetProvider(req.Provider, req.Model, forceStub)
	if err != nil {
		return ChatResult{}, err
	}

	result, err := provider.Generate(prompt)
	if err != nil {
		return ChatResult{}, err
	}

	if len(citations) > 0 {
		result.Answer += "\n\nCitations: " + joinStrings(citations, " ")
	}

	return ChatResult{
		Response:        result,
		Citations:       citations,
		RAGUsed:         ragUsed,
		RetrievedChunks: retrievedChunks,
	}, nil
}

func joinStrings(items []string, sep string) string {
	if len(items) == 0 {
		return ""
	}
	result := items[0]
	for _, item := range items[1:] {
		result += sep + item
	}
	return result
}
