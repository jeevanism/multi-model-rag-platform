import { useEffect, useMemo, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import './index.css'

type Provider = 'gemini' | 'openai'
type Mode = 'json' | 'stream'

type RetrievedChunk = {
  document_id: number
  chunk_id: number
  chunk_index: number
  title: string
  content: string
  score: number
}

type ChatApiResponse = {
  answer: string
  provider: string
  model: string
  latency_ms: number
  tokens_in: number | null
  tokens_out: number | null
  cost_usd: number | null
  citations: string[]
  rag_used: boolean
  retrieved_chunks: RetrievedChunk[] | null
}

type ChatMessage = {
  id: string
  role: 'user' | 'assistant'
  text: string
  meta?: {
    provider?: string
    model?: string
    latencyMs?: number
    costUsd?: number | null
    citations?: string[]
    ragUsed?: boolean
    retrievedChunks?: RetrievedChunk[] | null
    mode?: Mode
  }
}

type StreamStartEvent = { provider: string; model: string }
type StreamTokenEvent = { text: string }
type StreamEndEvent = {
  answer: string
  latency_ms: number
  tokens_in: number | null
  tokens_out: number | null
  cost_usd: number | null
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8000'

function App() {
  const [provider, setProvider] = useState<Provider>('gemini')
  const [model, setModel] = useState('')
  const [mode, setMode] = useState<Mode>('json')
  const [rag, setRag] = useState(true)
  const [debug, setDebug] = useState(true)
  const [topK, setTopK] = useState(2)
  const [input, setInput] = useState('What is the capital of France?')
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [requestId, setRequestId] = useState<string | null>(null)
  const scrollRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    scrollRef.current?.scrollTo({
      top: scrollRef.current.scrollHeight,
      behavior: 'smooth',
    })
  }, [messages])

  const providerPlaceholder = useMemo(() => {
    return provider === 'gemini' ? 'gemini-2.5-flash' : 'gpt-4.1-mini'
  }, [provider])

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!input.trim() || isLoading) {
      return
    }

    setError(null)
    setIsLoading(true)

    const message = input.trim()
    const requestPayload = {
      message,
      provider,
      model: model.trim() || null,
      rag,
      top_k: topK,
      debug,
    }

    const userMessage: ChatMessage = {
      id: crypto.randomUUID(),
      role: 'user',
      text: message,
    }

    setMessages((prev) => [...prev, userMessage])
    setInput('')

    try {
      if (mode === 'stream') {
        await submitStream(requestPayload)
      } else {
        await submitJson(requestPayload)
      }
    } catch (submitError) {
      const text =
        submitError instanceof Error ? submitError.message : 'Unexpected error while sending request.'
      setError(text)
    } finally {
      setIsLoading(false)
    }
  }

  async function submitJson(requestPayload: {
    message: string
    provider: Provider
    model: string | null
    rag: boolean
    top_k: number
    debug: boolean
  }) {
    const response = await fetch(`${API_BASE_URL}/chat`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestPayload),
    })

    setRequestId(response.headers.get('x-request-id'))

    if (!response.ok) {
      throw new Error(await readError(response))
    }

    const body = (await response.json()) as ChatApiResponse
    const assistantMessage: ChatMessage = {
      id: crypto.randomUUID(),
      role: 'assistant',
      text: body.answer,
      meta: {
        provider: body.provider,
        model: body.model,
        latencyMs: body.latency_ms,
        costUsd: body.cost_usd,
        citations: body.citations,
        ragUsed: body.rag_used,
        retrievedChunks: body.retrieved_chunks,
        mode: 'json',
      },
    }

    setMessages((prev) => [...prev, assistantMessage])
  }

  async function submitStream(requestPayload: {
    message: string
    provider: Provider
    model: string | null
    rag: boolean
    top_k: number
    debug: boolean
  }) {
    const response = await fetch(`${API_BASE_URL}/chat/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestPayload),
    })

    setRequestId(response.headers.get('x-request-id'))

    if (!response.ok) {
      throw new Error(await readError(response))
    }
    if (!response.body) {
      throw new Error('Streaming response body is not available.')
    }

    const assistantId = crypto.randomUUID()
    setMessages((prev) => [
      ...prev,
      {
        id: assistantId,
        role: 'assistant',
        text: '',
        meta: {
          citations: [],
          ragUsed: false,
          mode: 'stream',
        },
      },
    ])

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    let currentEvent = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })

      let separatorIndex = buffer.indexOf('\n\n')
      while (separatorIndex !== -1) {
        const rawEvent = buffer.slice(0, separatorIndex)
        buffer = buffer.slice(separatorIndex + 2)
        handleSseBlock(rawEvent)
        separatorIndex = buffer.indexOf('\n\n')
      }
    }

    function handleSseBlock(block: string) {
      if (!block.trim()) return

      let dataLine = ''
      for (const line of block.split('\n')) {
        if (line.startsWith('event:')) {
          currentEvent = line.replace('event:', '').trim()
        }
        if (line.startsWith('data:')) {
          dataLine = line.replace('data:', '').trim()
        }
      }

      if (!dataLine) return

      const parsed = JSON.parse(dataLine) as StreamStartEvent | StreamTokenEvent | StreamEndEvent

      if (currentEvent === 'start') {
        const start = parsed as StreamStartEvent
        setMessages((prev) =>
          prev.map((msg) =>
            msg.id === assistantId
              ? {
                  ...msg,
                  meta: {
                    ...msg.meta,
                    provider: start.provider,
                    model: start.model,
                    mode: 'stream',
                  },
                }
              : msg,
          ),
        )
        return
      }

      if (currentEvent === 'token') {
        const token = parsed as StreamTokenEvent
        setMessages((prev) =>
          prev.map((msg) =>
            msg.id === assistantId
              ? {
                  ...msg,
                  text: `${msg.text}${msg.text ? ' ' : ''}${token.text}`,
                }
              : msg,
          ),
        )
        return
      }

      if (currentEvent === 'end') {
        const end = parsed as StreamEndEvent
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
                    mode: 'stream',
                  },
                }
              : msg,
          ),
        )
      }
    }
  }

  return (
    <div className="app-shell">
      <aside className="control-panel" aria-label="Chat controls">
        <div className="panel-card">
          <p className="eyebrow">Multi-Model RAG UI</p>
          <h1>Iteration 12 Frontend</h1>
          <p className="panel-copy">
            Chat against your local FastAPI backend with provider selection, streaming mode,
            citations, and debug retrieval previews.
          </p>
        </div>

        <div className="panel-card stack">
          <label>
            <span>API Base URL</span>
            <input value={API_BASE_URL} disabled />
          </label>

          <label>
            <span>Provider</span>
            <select value={provider} onChange={(e) => setProvider(e.target.value as Provider)}>
              <option value="gemini">Gemini</option>
              <option value="openai">OpenAI</option>
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
            <select value={mode} onChange={(e) => setMode(e.target.value as Mode)}>
              <option value="json">/chat (JSON)</option>
              <option value="stream">/chat/stream (SSE)</option>
            </select>
          </label>

          <div className="toggle-grid">
            <label className="toggle">
              <input type="checkbox" checked={rag} onChange={(e) => setRag(e.target.checked)} />
              <span>RAG</span>
            </label>
            <label className="toggle">
              <input type="checkbox" checked={debug} onChange={(e) => setDebug(e.target.checked)} />
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
      </aside>

      <main className="chat-stage">
        <header className="chat-header">
          <div>
            <p className="eyebrow">Local API</p>
            <h2>Chat Workspace</h2>
          </div>
          <button
            type="button"
            className="ghost-button"
            onClick={() => {
              setMessages([])
              setError(null)
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
                Tip: Keep <strong>RAG</strong> and <strong>Debug</strong> enabled to see citations and
                retrieved chunks.
              </p>
            </div>
          ) : (
            messages.map((message) => (
              <article
                key={message.id}
                className={`message-card ${message.role === 'assistant' ? 'assistant' : 'user'}`}
              >
                <header className="message-meta">
                  <span className="role-tag">{message.role}</span>
                  {message.meta?.provider ? (
                    <span className="meta-pill">
                      {message.meta.provider} · {message.meta.model}
                    </span>
                  ) : null}
                  {typeof message.meta?.latencyMs === 'number' ? (
                    <span className="meta-pill">{message.meta.latencyMs} ms</span>
                  ) : null}
                  {typeof message.meta?.costUsd === 'number' ? (
                    <span className="meta-pill">${message.meta.costUsd.toFixed(4)}</span>
                  ) : null}
                  {message.meta?.mode ? <span className="meta-pill">{message.meta.mode}</span> : null}
                  {message.meta?.ragUsed ? <span className="meta-pill">rag=true</span> : null}
                </header>

                <pre className="message-text">{message.text}</pre>

                {message.meta?.citations && message.meta.citations.length > 0 ? (
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

                {message.meta?.retrievedChunks && message.meta.retrievedChunks.length > 0 ? (
                  <section className="subpanel">
                    <h3>Retrieved Chunks (Debug)</h3>
                    <div className="chunk-list">
                      {message.meta.retrievedChunks.map((chunk) => (
                        <article key={chunk.chunk_id} className="chunk-card">
                          <div className="chunk-card-header">
                            <strong>{chunk.title}</strong>
                            <span>
                              chunk {chunk.chunk_index} · score {chunk.score.toFixed(3)}
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
            {error ? <p className="error-text">{error}</p> : <span className="hint">Ready</span>}
            <button type="submit" disabled={isLoading || !input.trim()}>
              {isLoading ? 'Sending...' : mode === 'stream' ? 'Send (SSE)' : 'Send'}
            </button>
          </div>
        </form>
      </main>
    </div>
  )
}

async function readError(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as { detail?: string }
    return body.detail ?? `Request failed with status ${response.status}`
  } catch {
    return `Request failed with status ${response.status}`
  }
}

export default App
