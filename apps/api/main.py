from fastapi import FastAPI, HTTPException

from apps.api.schemas.chat import ChatRequest, ChatResponse
from apps.api.services.chat import generate_chat_response
from packages.llm.router import UnsupportedProviderError

app = FastAPI(title="Multi-Model RAG API", version="0.1.0")


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok", "service": "api"}


@app.post("/chat", response_model=ChatResponse)
def chat(request: ChatRequest) -> ChatResponse:
    try:
        return generate_chat_response(request)
    except UnsupportedProviderError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
