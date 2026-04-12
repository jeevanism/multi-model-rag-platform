package service

import (
	"errors"

	"multi-model-rag-platform/internal/llm"
)

var ErrRAGNotImplemented = errors.New("RAG is not implemented in the Go backend yet")

type ChatParams struct {
	Message  string
	Provider string
	Model    *string
	RAG      bool
}

type ChatService struct {
	router llm.Router
}

func NewChatService(router llm.Router) ChatService {
	return ChatService{router: router}
}

func (s ChatService) GenerateChatResponse(req ChatParams, forceStub bool) (llm.Response, error) {
	if req.RAG {
		return llm.Response{}, ErrRAGNotImplemented
	}

	provider, err := s.router.GetProvider(req.Provider, req.Model, forceStub)
	if err != nil {
		return llm.Response{}, err
	}

	result, err := provider.Generate(req.Message)
	if err != nil {
		return llm.Response{}, err
	}

	return result, nil
}
