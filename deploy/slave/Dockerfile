# ============================================================
# Stage 1: Build Go Backend (no frontend — slave is API-only)
# ============================================================
FROM golang:1.23-bookworm AS backend-builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
ENV GOSUMDB=off

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -linkmode external -extldflags '-static' -X github.com/quantumclaw/quantumclaw/common.Version=${VERSION}" \
    -o quantumclaw .

# ============================================================
# Stage 2: Runtime
# ============================================================
FROM alpine:3.19

LABEL maintainer="QuantumClaw" \
      description="QuantumClaw Slave Node - AI API Relay"

RUN apk add --no-cache ca-certificates tzdata dumb-init su-exec && \
    addgroup -g 1000 quantumclaw && \
    adduser -u 1000 -G quantumclaw -s /bin/sh -D quantumclaw

WORKDIR /app

COPY --from=backend-builder /app/quantumclaw .
RUN mkdir -p /app/logs && chown -R quantumclaw:quantumclaw /app/logs

USER quantumclaw
EXPOSE 3666

ENV PORT=3666 GIN_MODE=release LOG_DIR=/app/logs

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:3666/health || exit 1

ENTRYPOINT ["dumb-init", "--"]
CMD ["./quantumclaw"]
