import { useEffect, useMemo, useRef, useState } from "react";
import type { FormEvent } from "react";
import "./index.css";

type Provider = "gemini" | "openai" | "grok";
type Mode = "json" | "stream";
type View = "chat" | "evals";

type RetrievedChunk = {
  document_id: number;
  chunk_id: number;
  chunk_index: number;
  title: string;
  content: string;
  score: number;
};

type ChatApiResponse = {
  answer: string;
  provider: string;
  model: string;
  latency_ms: number;
  tokens_in: number | null;
  tokens_out: number | null;
  cost_usd: number | null;
  citations: string[];
  rag_used: boolean;
  retrieved_chunks: RetrievedChunk[] | null;
};

type ChatMessage = {
  id: string;
  role: "user" | "assistant";
  text: string;
  meta?: {
    provider?: string;
    model?: string;
    latencyMs?: number;
    costUsd?: number | null;
    citations?: string[];
    ragUsed?: boolean;
    retrievedChunks?: RetrievedChunk[] | null;
    mode?: Mode;
  };
};

type StreamStartEvent = { provider: string; model: string };
type StreamTokenEvent = { text: string };
type StreamEndEvent = {
  answer: string;
  latency_ms: number;
  tokens_in: number | null;
  tokens_out: number | null;
  cost_usd: number | null;
};

type EvalRunSummaryItem = {
  id: number;
  dataset_name: string;
  provider: string;
  model: string | null;
  total_cases: number;
  passed_cases: number;
  avg_latency_ms: number | null;
  created_at: string;
};

type EvalRunCaseItem = {
  id: number;
  case_id: string;
  question: string;
  passed: boolean;
  latency_ms: number;
  correctness_score: number | null;
  groundedness_score: number | null;
  hallucination_score: number | null;
  citations: string[];
  error: string | null;
};

type EvalRunDetail = {
  run: EvalRunSummaryItem;
  cases: EvalRunCaseItem[];
};

type DemoAuthStatus = {
  unlocked: boolean;
  unlock_enabled: boolean;
};

const MODEL_PRESETS: Record<Provider, string[]> = {
  gemini: [
    "gemini-3.1-pro-preview",
    "gemini-3-pro-preview",
    "gemini-3-flash-preview",
    "gemini-2.5-pro",
    "gemini-2.5-flash",
    "gemini-2.5-flash-lite",
  ],
  openai: ["gpt-4.1-mini", "gpt-4.1", "gpt-4o-mini"],
  grok: [
    "grok-4-fast-non-reasoning",
    "grok-4-fast-reasoning",
    "grok-4-1-fast-non-reasoning",
    "grok-4-1-fast-reasoning",
    "grok-code-fast-1",
    "grok-3-mini",
  ],
};

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8000";
const CLOUD_SQL_PAUSED_MESSAGE =
  "Cloud SQL is stopped to save billing. Please contact jeevz.dev@gmail.com to enable it.";

function App() {
  const [view, setView] = useState<View>("chat");
  const [provider, setProvider] = useState<Provider>("gemini");
  const [model, setModel] = useState("");
  const [mode, setMode] = useState<Mode>("json");
  const [rag, setRag] = useState(true);
  const [debug, setDebug] = useState(true);
  const [topK, setTopK] = useState(2);
  const [input, setInput] = useState("What is the capital of France?");
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [requestId, setRequestId] = useState<string | null>(null);
  const [evalRuns, setEvalRuns] = useState<EvalRunSummaryItem[]>([]);
  const [selectedEvalRunId, setSelectedEvalRunId] = useState<number | null>(
    null,
  );
  const [evalRunDetail, setEvalRunDetail] = useState<EvalRunDetail | null>(
    null,
  );
  const [evalsLoading, setEvalsLoading] = useState(false);
  const [evalsError, setEvalsError] = useState<string | null>(null);
  const [realModeUnlocked, setRealModeUnlocked] = useState(false);
  const [realModeAvailable, setRealModeAvailable] = useState(false);
  const [unlockModalOpen, setUnlockModalOpen] = useState(false);
  const [unlockPassword, setUnlockPassword] = useState("");
  const [unlocking, setUnlocking] = useState(false);
  const [unlockError, setUnlockError] = useState<string | null>(null);
  const scrollRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    scrollRef.current?.scrollTo({
      top: scrollRef.current.scrollHeight,
      behavior: "smooth",
    });
  }, [messages]);

  useEffect(() => {
    if (view === "evals") {
      void loadEvalRuns();
    }
  }, [view]);

  useEffect(() => {
    void loadDemoAuthStatus();
  }, []);

  useEffect(() => {
    if (selectedEvalRunId !== null) {
      void loadEvalRunDetail(selectedEvalRunId);
    }
  }, [selectedEvalRunId]);

  const providerPlaceholder = useMemo(() => {
    if (provider === "gemini") return "gemini-2.5-flash";
    if (provider === "grok") return "grok-3-mini";
    return "gpt-4.1-mini";
  }, [provider]);

  const selectedPresetValue = useMemo(() => {
    if (!model.trim()) return "";
    return MODEL_PRESETS[provider].includes(model.trim())
      ? model.trim()
      : "__custom__";
  }, [model, provider]);

  async function loadEvalRuns() {
    setEvalsLoading(true);
    setEvalsError(null);
    try {
      const response = await fetch(`${API_BASE_URL}/evals/runs`, {
        credentials: "include",
      });
      if (!response.ok) {
        throw new Error(await readError(response));
      }
      const runs = (await response.json()) as EvalRunSummaryItem[];
      setEvalRuns(runs);
      if (runs.length > 0 && selectedEvalRunId === null) {
        setSelectedEvalRunId(runs[0].id);
      }
    } catch (loadError) {
      setEvalsError(normalizeUiError(loadError, "Failed to load eval runs."));
    } finally {
      setEvalsLoading(false);
    }
  }

  async function loadEvalRunDetail(evalRunId: number) {
    setEvalsLoading(true);
    setEvalsError(null);
    try {
      const response = await fetch(`${API_BASE_URL}/evals/runs/${evalRunId}`, {
        credentials: "include",
      });
      if (!response.ok) {
        throw new Error(await readError(response));
      }
      const detail = (await response.json()) as EvalRunDetail;
      setEvalRunDetail(detail);
    } catch (loadError) {
      setEvalsError(
        normalizeUiError(loadError, "Failed to load eval run detail."),
      );
    } finally {
      setEvalsLoading(false);
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!input.trim() || isLoading) return;

    setError(null);
    setIsLoading(true);

    const message = input.trim();
    const requestPayload = {
      message,
      provider,
      model: model.trim() || null,
      rag,
      top_k: topK,
      debug,
    };

    setMessages((prev) => [
      ...prev,
      { id: crypto.randomUUID(), role: "user", text: message },
    ]);
    setInput("");

    try {
      if (mode === "stream") {
        await submitStream(requestPayload);
      } else {
        await submitJson(requestPayload);
      }
    } catch (submitError) {
      setError(
        normalizeUiError(
          submitError,
          "Unexpected error while sending request.",
        ),
      );
    } finally {
      setIsLoading(false);
    }
  }

  async function submitJson(requestPayload: {
    message: string;
    provider: Provider;
    model: string | null;
    rag: boolean;
    top_k: number;
    debug: boolean;
  }) {
    const response = await fetch(`${API_BASE_URL}/chat`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(requestPayload),
    });
    setRequestId(response.headers.get("x-request-id"));
    if (!response.ok) throw new Error(await readError(response));

    const body = (await response.json()) as ChatApiResponse;
    setMessages((prev) => [
      ...prev,
      {
        id: crypto.randomUUID(),
        role: "assistant",
        text: body.answer,
        meta: {
          provider: body.provider,
          model: body.model,
          latencyMs: body.latency_ms,
          costUsd: body.cost_usd,
          citations: body.citations,
          ragUsed: body.rag_used,
          retrievedChunks: body.retrieved_chunks,
          mode: "json",
        },
      },
    ]);
  }

  async function submitStream(requestPayload: {
    message: string;
    provider: Provider;
    model: string | null;
    rag: boolean;
    top_k: number;
    debug: boolean;
  }) {
    const response = await fetch(`${API_BASE_URL}/chat/stream`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(requestPayload),
    });
    setRequestId(response.headers.get("x-request-id"));
    if (!response.ok) throw new Error(await readError(response));
    if (!response.body)
      throw new Error("Streaming response body is not available.");

    const assistantId = crypto.randomUUID();
    setMessages((prev) => [
      ...prev,
      {
        id: assistantId,
        role: "assistant",
        text: "",
        meta: { citations: [], ragUsed: false, mode: "stream" },
      },
    ]);

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";
    let currentEvent = "";

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      let separatorIndex = buffer.indexOf("\n\n");
      while (separatorIndex !== -1) {
        const rawEvent = buffer.slice(0, separatorIndex);
        buffer = buffer.slice(separatorIndex + 2);
        handleSseBlock(rawEvent);
        separatorIndex = buffer.indexOf("\n\n");
      }
    }

    function handleSseBlock(block: string) {
      if (!block.trim()) return;
      let dataLine = "";
      for (const line of block.split("\n")) {
        if (line.startsWith("event:"))
          currentEvent = line.replace("event:", "").trim();
        if (line.startsWith("data:"))
          dataLine = line.replace("data:", "").trim();
      }
      if (!dataLine) return;
      const parsed = JSON.parse(dataLine) as
        | StreamStartEvent
        | StreamTokenEvent
        | StreamEndEvent;

      if (currentEvent === "start") {
        const start = parsed as StreamStartEvent;
        setMessages((prev) =>
          prev.map((msg) =>
            msg.id === assistantId
              ? {
                  ...msg,
                  meta: {
                    ...msg.meta,
                    provider: start.provider,
                    model: start.model,
                    mode: "stream",
                  },
                }
              : msg,
          ),
        );
        return;
      }
      if (currentEvent === "token") {
        const token = parsed as StreamTokenEvent;
        setMessages((prev) =>
          prev.map((msg) =>
            msg.id === assistantId
              ? {
                  ...msg,
                  text: `${msg.text}${msg.text ? " " : ""}${token.text}`,
                }
              : msg,
          ),
        );
        return;
      }
      if (currentEvent === "end") {
        const end = parsed as StreamEndEvent;
        setMessages((prev) =>
          prev.map((msg) =>
            msg.id === assistantId
              ? {
                  ...msg,
                  text: end.answer,
                  meta: {
                    ...msg.meta,
                    latencyMs: end.latency_ms,
                    costUsd: end.cost_usd,
                    mode: "stream",
                  },
                }
              : msg,
          ),
        );
      }
    }
  }

  async function loadDemoAuthStatus() {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/demo-status`, {
        credentials: "include",
      });
      if (!response.ok) return;
      const status = (await response.json()) as DemoAuthStatus;
      setRealModeUnlocked(status.unlocked);
      setRealModeAvailable(status.unlock_enabled);
    } catch {
      // Non-blocking; chat should still function in public stub mode.
    }
  }

  async function unlockRealMode() {
    setUnlocking(true);
    setUnlockError(null);
    try {
      const response = await fetch(`${API_BASE_URL}/auth/demo-unlock`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ password: unlockPassword }),
      });
      if (!response.ok) {
        throw new Error(await readError(response));
      }
      const status = (await response.json()) as DemoAuthStatus;
      setRealModeUnlocked(status.unlocked);
      setRealModeAvailable(status.unlock_enabled);
      setUnlockModalOpen(false);
      setUnlockPassword("");
    } catch (err) {
      setUnlockError(err instanceof Error ? err.message : "Unlock failed");
    } finally {
      setUnlocking(false);
    }
  }

  async function lockRealMode() {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/demo-lock`, {
        method: "POST",
        credentials: "include",
      });
      if (response.ok) {
        const status = (await response.json()) as DemoAuthStatus;
        setRealModeUnlocked(status.unlocked);
        setRealModeAvailable(status.unlock_enabled);
      }
    } catch {
      setRealModeUnlocked(false);
    }
  }

  return (
    <>
      <div className="app-shell">
        <aside className="control-panel" aria-label="Controls">
          <div className="panel-card">
            <p className="eyebrow">Multi-Model RAG UI</p>
            <h1>Frontend</h1>
            <p className="panel-copy">
              Chat + eval dashboard in one workspace. Use the tabs below to
              switch between runtime testing and eval review.
            </p>
            <div
              className="tab-switch"
              role="tablist"
              aria-label="Workspace tabs"
            >
              <button
                type="button"
                role="tab"
                aria-selected={view === "chat"}
                className={view === "chat" ? "tab active" : "tab"}
                onClick={() => setView("chat")}
              >
                Chat
              </button>
              <button
                type="button"
                role="tab"
                aria-selected={view === "evals"}
                className={view === "evals" ? "tab active" : "tab"}
                onClick={() => setView("evals")}
              >
                Evals
              </button>
            </div>
          </div>

          {view === "chat" ? (
            <div className="panel-card stack">
              <label>
                <span>API Base URL</span>
                <input value={API_BASE_URL} disabled />
              </label>
              <div className="auth-status-card">
                <div>
                  <span className="auth-status-label">Real LLM Mode</span>
                  <strong
                    className={
                      realModeUnlocked ? "auth-status unlocked" : "auth-status"
                    }
                  >
                    {realModeUnlocked ? "Unlocked" : "Public Stub Mode"}
                  </strong>
                </div>
                {realModeUnlocked ? (
                  <button
                    type="button"
                    className="ghost-button"
                    onClick={() => void lockRealMode()}
                  >
                    Lock
                  </button>
                ) : realModeAvailable ? (
                  <button
                    type="button"
                    className="ghost-button"
                    onClick={() => {
                      setUnlockError(null);
                      setUnlockModalOpen(true);
                    }}
                  >
                    Unlock
                  </button>
                ) : (
                  <span className="hint">Disabled</span>
                )}
              </div>
              <label>
                <span>Provider</span>
                <select
                  value={provider}
                  onChange={(e) => setProvider(e.target.value as Provider)}
                >
                  <option value="gemini">Gemini</option>
                  <option value="openai">OpenAI</option>
                  <option value="grok">Grok</option>
                </select>
              </label>
              <label>
                <span>Model Preset</span>
                <select
                  value={selectedPresetValue}
                  onChange={(e) => {
                    const next = e.target.value;
                    if (next === "__custom__") return;
                    setModel(next);
                  }}
                >
                  <option value="">Default ({providerPlaceholder})</option>
                  {MODEL_PRESETS[provider].map((presetModel) => (
                    <option key={presetModel} value={presetModel}>
                      {presetModel}
                    </option>
                  ))}
                  <option value="__custom__">Custom (use field below)</option>
                </select>
              </label>
              <label>
                <span>Model (optional override)</span>
                <input
                  value={model}
                  onChange={(e) => setModel(e.target.value)}
                  placeholder={providerPlaceholder}
                />
              </label>
              <label>
                <span>Mode</span>
                <select
                  value={mode}
                  onChange={(e) => setMode(e.target.value as Mode)}
                >
                  <option value="json">/chat (JSON)</option>
                  <option value="stream">/chat/stream (SSE)</option>
                </select>
              </label>
              <div className="toggle-grid">
                <label className="toggle">
                  <input
                    type="checkbox"
                    checked={rag}
                    onChange={(e) => setRag(e.target.checked)}
                  />
                  <span>RAG</span>
                </label>
                <label className="toggle">
                  <input
                    type="checkbox"
                    checked={debug}
                    onChange={(e) => setDebug(e.target.checked)}
                  />
                  <span>Debug</span>
                </label>
              </div>
              <label>
                <span>Top K</span>
                <input
                  type="number"
                  min={1}
                  max={10}
                  value={topK}
                  onChange={(e) => setTopK(Number(e.target.value))}
                  disabled={!rag}
                />
              </label>
              {requestId ? (
                <div className="request-id-box">
                  <span>Last Request ID</span>
                  <code>{requestId}</code>
                </div>
              ) : null}
            </div>
          ) : (
            <div className="panel-card stack">
              <div className="row-between">
                <span>Recent Eval Runs</span>
                <button
                  type="button"
                  className="ghost-button"
                  onClick={() => void loadEvalRuns()}
                >
                  Refresh
                </button>
              </div>
              {evalsLoading && evalRuns.length === 0 ? (
                <p className="hint">Loading eval runs...</p>
              ) : null}
              {evalsError ? <p className="error-text">{evalsError}</p> : null}
              <div className="run-list">
                {evalRuns.map((run) => {
                  const passRate =
                    run.total_cases > 0
                      ? Math.round((run.passed_cases / run.total_cases) * 100)
                      : 0;
                  return (
                    <button
                      key={run.id}
                      type="button"
                      className={
                        selectedEvalRunId === run.id
                          ? "run-item active"
                          : "run-item"
                      }
                      onClick={() => setSelectedEvalRunId(run.id)}
                    >
                      <div className="run-item-top">
                        <strong>Run #{run.id}</strong>
                        <span>{passRate}% pass</span>
                      </div>
                      <div className="run-item-meta">
                        <span>{run.dataset_name}</span>
                        <span>
                          {run.provider}
                          {run.model ? ` · ${run.model}` : ""}
                        </span>
                      </div>
                    </button>
                  );
                })}
              </div>
            </div>
          )}
        </aside>

        <main className="chat-stage">
          {view === "chat" ? (
            <>
              <header className="chat-header">
                <div>
                  <p className="eyebrow">Cloud API</p>
                  <h2>Chat Workspace</h2>
                </div>
                <button
                  type="button"
                  className="ghost-button"
                  onClick={() => {
                    setMessages([]);
                    setError(null);
                  }}
                >
                  Clear
                </button>
              </header>
              <div className="chat-log" ref={scrollRef}>
                {messages.length === 0 ? (
                  <div className="empty-state">
                    <p>Send a message to test `/chat` or `/chat/stream`.</p>

                    <p className="hint">
                      Tip: The current ingested document contains content about
                      capitals (France/Germany). To demonstrate RAG behavior on
                      an out-of-scope query, ask a question unrelated to the
                      ingested document (for example, What is 2+2?). This helps
                      us observe retrieval/citation behavior and why evals are
                      needed to detect hallucination and grounding failures.
                    </p>
                  </div>
                ) : (
                  messages.map((message) => (
                    <article
                      key={message.id}
                      className={`message-card ${message.role === "assistant" ? "assistant" : "user"}`}
                    >
                      <header className="message-meta">
                        <span className="role-tag">{message.role}</span>
                        {message.meta?.provider ? (
                          <span className="meta-pill">
                            {message.meta.provider} · {message.meta.model}
                          </span>
                        ) : null}
                        {typeof message.meta?.latencyMs === "number" ? (
                          <span className="meta-pill">
                            {message.meta.latencyMs} ms
                          </span>
                        ) : null}
                        {typeof message.meta?.costUsd === "number" ? (
                          <span className="meta-pill">
                            ${message.meta.costUsd.toFixed(4)}
                          </span>
                        ) : null}
                        {message.meta?.mode ? (
                          <span className="meta-pill">{message.meta.mode}</span>
                        ) : null}
                        {message.meta?.ragUsed ? (
                          <span className="meta-pill">rag=true</span>
                        ) : null}
                      </header>
                      <pre className="message-text">{message.text}</pre>
                      {message.meta?.citations &&
                      message.meta.citations.length > 0 ? (
                        <section className="subpanel">
                          <h3>Citations</h3>
                          <ul>
                            {message.meta.citations.map((citation) => (
                              <li key={citation}>
                                <code>{citation}</code>
                              </li>
                            ))}
                          </ul>
                        </section>
                      ) : null}
                      {message.meta?.retrievedChunks &&
                      message.meta.retrievedChunks.length > 0 ? (
                        <section className="subpanel">
                          <h3>Retrieved Chunks (Debug)</h3>
                          <div className="chunk-list">
                            {message.meta.retrievedChunks.map((chunk) => (
                              <article
                                key={chunk.chunk_id}
                                className="chunk-card"
                              >
                                <div className="chunk-card-header">
                                  <strong>{chunk.title}</strong>
                                  <span>
                                    chunk {chunk.chunk_index} · score{" "}
                                    {chunk.score.toFixed(3)}
                                  </span>
                                </div>
                                <p>{chunk.content}</p>
                              </article>
                            ))}
                          </div>
                        </section>
                      ) : null}
                    </article>
                  ))
                )}
              </div>
              <form className="composer" onSubmit={handleSubmit}>
                <label htmlFor="message-input" className="sr-only">
                  Message
                </label>
                <textarea
                  id="message-input"
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  placeholder="Ask something..."
                  rows={4}
                  disabled={isLoading}
                />
                <div className="composer-footer">
                  {error ? (
                    <p className="error-text">{error}</p>
                  ) : (
                    <span className="hint">Ready</span>
                  )}
                  <button type="submit" disabled={isLoading || !input.trim()}>
                    {isLoading
                      ? "Sending..."
                      : mode === "stream"
                        ? "Send (SSE)"
                        : "Send"}
                  </button>
                </div>
              </form>
            </>
          ) : (
            <>
              <header className="chat-header">
                <div>
                  <p className="eyebrow">Evals</p>
                  <h2>Evaluation Dashboard</h2>
                </div>
                <button
                  type="button"
                  className="ghost-button"
                  onClick={() => void loadEvalRuns()}
                >
                  Refresh Runs
                </button>
              </header>
              <div className="chat-log eval-dashboard">
                {evalsError ? <p className="error-text">{evalsError}</p> : null}
                {!evalRunDetail && !evalsLoading ? (
                  <p className="hint">Select an eval run to inspect details.</p>
                ) : null}
                {evalRunDetail ? (
                  <div className="eval-detail">
                    <section className="subpanel eval-summary-grid">
                      <div className="metric-card">
                        <span>Run ID</span>
                        <strong>#{evalRunDetail.run.id}</strong>
                      </div>
                      <div className="metric-card">
                        <span>Dataset</span>
                        <strong>{evalRunDetail.run.dataset_name}</strong>
                      </div>
                      <div className="metric-card">
                        <span>Pass</span>
                        <strong>
                          {evalRunDetail.run.passed_cases}/
                          {evalRunDetail.run.total_cases}
                        </strong>
                      </div>
                      <div className="metric-card">
                        <span>Latency Avg</span>
                        <strong>
                          {evalRunDetail.run.avg_latency_ms ?? 0} ms
                        </strong>
                      </div>
                    </section>
                    <section className="subpanel">
                      <h3>Cases</h3>
                      <div className="eval-table-wrap">
                        <table className="eval-table">
                          <thead>
                            <tr>
                              <th>Case</th>
                              <th>Status</th>
                              <th>Latency</th>
                              <th>Correct</th>
                              <th>Grounded</th>
                              <th>Halluc.</th>
                            </tr>
                          </thead>
                          <tbody>
                            {evalRunDetail.cases.map((row) => (
                              <tr key={row.id}>
                                <td>
                                  <div className="case-cell">
                                    <strong>{row.case_id}</strong>
                                    <span>{row.question}</span>
                                  </div>
                                </td>
                                <td>
                                  <span
                                    className={
                                      row.passed ? "status-pass" : "status-fail"
                                    }
                                  >
                                    {row.passed ? "pass" : "fail"}
                                  </span>
                                </td>
                                <td>{row.latency_ms} ms</td>
                                <td>{formatScore(row.correctness_score)}</td>
                                <td>{formatScore(row.groundedness_score)}</td>
                                <td>{formatScore(row.hallucination_score)}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </section>
                  </div>
                ) : null}
              </div>
            </>
          )}
        </main>
      </div>
      {unlockModalOpen ? (
        <div
          className="modal-backdrop"
          role="dialog"
          aria-modal="true"
          aria-labelledby="unlock-modal-title"
        >
          <div className="modal-card">
            <div className="row-between">
              <h3 id="unlock-modal-title">Unlock Real Mode</h3>
              <button
                type="button"
                className="ghost-button"
                onClick={() => setUnlockModalOpen(false)}
              >
                Close
              </button>
            </div>
            <p className="hint">
              Enter the demo password to enable real provider calls for this
              session.
            </p>
            <label>
              <span>Password</span>
              <input
                type="password"
                value={unlockPassword}
                onChange={(e) => setUnlockPassword(e.target.value)}
                autoFocus
              />
            </label>
            {unlockError ? <p className="error-text">{unlockError}</p> : null}
            <div className="modal-actions">
              <button
                type="button"
                className="ghost-button"
                onClick={() => setUnlockModalOpen(false)}
                disabled={unlocking}
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={() => void unlockRealMode()}
                disabled={unlocking || !unlockPassword.trim()}
              >
                {unlocking ? "Unlocking..." : "Unlock Real Mode"}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}

function formatScore(value: number | null): string {
  return value == null ? "—" : value.toFixed(2);
}

async function readError(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as { detail?: string };
    return body.detail ?? `Request failed with status ${response.status}`;
  } catch {
    return `Request failed with status ${response.status}`;
  }
}

function normalizeUiError(error: unknown, fallback: string): string {
  if (error instanceof Error) {
    if (isLikelyNetworkFailure(error.message)) {
      return CLOUD_SQL_PAUSED_MESSAGE;
    }
    return error.message;
  }
  return fallback;
}

function isLikelyNetworkFailure(message: string): boolean {
  const normalized = message.toLowerCase();
  return (
    normalized.includes("failed to fetch") ||
    normalized.includes("networkerror when attempting to fetch resource") ||
    normalized.includes("network request failed") ||
    normalized.includes("load failed")
  );
}

export default App;
