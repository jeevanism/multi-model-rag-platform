from fastapi import FastAPI, HTTPException
from fastapi.responses import StreamingResponse

from apps.api.schemas.chat import ChatRequest, ChatResponse
from apps.api.services.chat import generate_chat_response, stream_chat_events
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


@app.post("/chat/stream")
def chat_stream(request: ChatRequest) -> StreamingResponse:
    try:
        event_stream = stream_chat_events(request)
    except UnsupportedProviderError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    return StreamingResponse(event_stream, media_type="text/event-stream")
