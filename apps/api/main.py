from fastapi import Depends, FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import StreamingResponse
from sqlalchemy.orm import Session

from apps.api import middleware
from apps.api.db import SessionLocal, get_db
from apps.api.schemas.chat import ChatRequest, ChatResponse
from apps.api.schemas.evals import EvalRunDetail, EvalRunSummaryItem
from apps.api.schemas.ingest import IngestTextRequest, IngestTextResponse
from apps.api.services.chat import generate_chat_response, stream_chat_events
from apps.api.services.evals import get_eval_run_detail, list_eval_runs
from apps.api.services.ingest import ingest_text_document
from apps.api.settings import settings
from packages.llm.router import UnsupportedProviderError
from packages.observability.logging import configure_logging
from packages.observability.request_context import get_request_id

app = FastAPI(title="Multi-Model RAG API", version="0.1.0")
configure_logging(settings.log_level)
app.middleware("http")(middleware.request_observability_middleware)
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.cors_allow_origins,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok", "service": "api", "request_id": get_request_id() or ""}


@app.post("/chat", response_model=ChatResponse)
def chat(request: ChatRequest) -> ChatResponse:
    db: Session | None = None
    try:
        if request.rag:
            db = SessionLocal()
        return generate_chat_response(request, db=db)
    except UnsupportedProviderError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    finally:
        if db is not None:
            db.close()


@app.post("/chat/stream")
def chat_stream(request: ChatRequest) -> StreamingResponse:
    try:
        event_stream = stream_chat_events(request)
    except UnsupportedProviderError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    return StreamingResponse(event_stream, media_type="text/event-stream")


@app.post("/ingest/text", response_model=IngestTextResponse)
def ingest_text(request: IngestTextRequest, db: Session = Depends(get_db)) -> IngestTextResponse:
    try:
        return ingest_text_document(db, request)
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc


@app.get("/evals/runs", response_model=list[EvalRunSummaryItem])
def eval_runs(db: Session = Depends(get_db)) -> list[EvalRunSummaryItem]:
    return list_eval_runs(db)


@app.get("/evals/runs/{eval_run_id}", response_model=EvalRunDetail)
def eval_run_detail(eval_run_id: int, db: Session = Depends(get_db)) -> EvalRunDetail:
    detail = get_eval_run_detail(db, eval_run_id)
    if detail is None:
        raise HTTPException(status_code=404, detail="Eval run not found")
    return detail
