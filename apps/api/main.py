from contextlib import ExitStack

from fastapi import Depends, FastAPI, HTTPException, Request, Response
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import StreamingResponse
from sqlalchemy.orm import Session

from apps.api import middleware
from apps.api.auth_demo import (
    clear_demo_unlock_cookie,
    demo_unlock_enabled,
    is_demo_unlocked,
    set_demo_unlock_cookie,
    validate_demo_password,
)
from apps.api.db import SessionLocal, get_db
from apps.api.schemas.auth import DemoUnlockRequest, DemoUnlockStatusResponse
from apps.api.schemas.chat import ChatRequest, ChatResponse
from apps.api.schemas.evals import EvalRunDetail, EvalRunSummaryItem
from apps.api.schemas.ingest import IngestTextRequest, IngestTextResponse
from apps.api.services.chat import generate_chat_response, stream_chat_events
from apps.api.services.evals import get_eval_run_detail, list_eval_runs
from apps.api.services.ingest import ingest_text_document
from apps.api.settings import settings
from packages.llm.router import UnsupportedProviderError
from packages.llm.router import override_provider_mode
from packages.observability.logging import configure_logging
from packages.observability.request_context import get_request_id
from packages.rag.embeddings import override_embedding_mode

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


@app.get("/")
def root() -> dict[str, object]:
    return {
        "service": "Multi-Model RAG API",
        "status": "ok",
        "endpoints": ["/health", "/chat", "/chat/stream", "/evals/runs"],
        "request_id": get_request_id() or "",
    }


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok", "service": "api", "request_id": get_request_id() or ""}


@app.post("/chat", response_model=ChatResponse)
def chat(request: ChatRequest, http_request: Request) -> ChatResponse:
    db: Session | None = None
    try:
        if request.rag:
            db = SessionLocal()
        unlocked = is_demo_unlocked(http_request)
        with ExitStack() as stack:
            if not unlocked:
                stack.enter_context(override_provider_mode("stub"))
                stack.enter_context(override_embedding_mode("stub"))
            return generate_chat_response(request, db=db)
    except UnsupportedProviderError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    finally:
        if db is not None:
            db.close()


@app.post("/chat/stream")
def chat_stream(request: ChatRequest, http_request: Request) -> StreamingResponse:
    try:
        unlocked = is_demo_unlocked(http_request)
        event_stream = stream_chat_events(request, force_stub=not unlocked)
    except UnsupportedProviderError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    return StreamingResponse(event_stream, media_type="text/event-stream")


@app.post("/ingest/text", response_model=IngestTextResponse)
def ingest_text(
    request: IngestTextRequest,
    http_request: Request,
    db: Session = Depends(get_db),
) -> IngestTextResponse:
    try:
        if is_demo_unlocked(http_request):
            return ingest_text_document(db, request)
        with override_embedding_mode("stub"):
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


@app.get("/auth/demo-status", response_model=DemoUnlockStatusResponse)
def demo_status(request: Request) -> DemoUnlockStatusResponse:
    return DemoUnlockStatusResponse(
        unlocked=is_demo_unlocked(request),
        unlock_enabled=demo_unlock_enabled(),
    )


@app.post("/auth/demo-unlock", response_model=DemoUnlockStatusResponse)
def demo_unlock(request: DemoUnlockRequest) -> Response:
    if not validate_demo_password(request.password):
        raise HTTPException(status_code=401, detail="Invalid demo password")

    response = Response(
        content=DemoUnlockStatusResponse(unlocked=True, unlock_enabled=True).model_dump_json(),
        media_type="application/json",
    )
    set_demo_unlock_cookie(response)
    return response


@app.post("/auth/demo-lock", response_model=DemoUnlockStatusResponse)
def demo_lock() -> Response:
    response = Response(
        content=DemoUnlockStatusResponse(
            unlocked=False, unlock_enabled=demo_unlock_enabled()
        ).model_dump_json(),
        media_type="application/json",
    )
    clear_demo_unlock_cookie(response)
    return response
