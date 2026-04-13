FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/multi-model-rag-api ./cmd/api

FROM debian:bookworm-slim AS runtime

WORKDIR /app

ENV PORT=8080

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --create-home --home-dir /home/appuser appuser

COPY --from=builder /out/multi-model-rag-api /app/multi-model-rag-api
COPY migrations ./migrations

USER appuser

EXPOSE 8080

CMD ["/app/multi-model-rag-api"]
