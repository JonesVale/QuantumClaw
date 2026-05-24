# ============================================================
# Stage 1: Build Frontend
# ============================================================
FROM node:22-alpine AS frontend-builder

WORKDIR /app

# 1. Install deps (cached when package.json unchanged)
COPY web/default/package*.json ./default/
RUN cd default && npm install --legacy-peer-deps && cd ..

# 2. Build frontend
COPY web/default ./default/
ARG VERSION=dev
RUN cd default && \
    VITE_REACT_APP_VERSION=${VERSION} npx rsbuild build && \
    cd .. && \
    mkdir -p /app/web/build && \
    mv default/dist /app/web/build/default

# ============================================================
# Stage 2: Build Go Backend
# ============================================================
FROM golang:1.23-bookworm AS backend-builder

# bookworm includes gcc, no need to install

WORKDIR /app

# Use Go proxy for faster downloads
ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
ENV GOSUMDB=off

# 1. Download modules (cached when go.mod/sum unchanged)
COPY go.mod go.sum ./
RUN go mod download

# 2. Copy source and built frontend
COPY . .
COPY --from=frontend-builder /app/web/build/default ./web/default/dist

# 3. Build
ARG VERSION=dev
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -linkmode external -extldflags '-static' -X github.com/quantumclaw/quantumclaw/common.Version=${VERSION}" \
    -o quantumclaw .

# ============================================================
# Stage 3: Runtime
# ============================================================
FROM alpine:3.19

LABEL maintainer="QuantumClaw" \
      description="QuantumClaw - AI API Gateway & Token Distribution Platform"

RUN apk add --no-cache ca-certificates tzdata dumb-init su-exec && \
    addgroup -g 1000 quantumclaw && \
    adduser -u 1000 -G quantumclaw -s /bin/sh -D quantumclaw

WORKDIR /app

COPY --from=backend-builder /app/quantumclaw .
COPY --from=backend-builder /app/web ./web

RUN mkdir -p /app/logs /app/data && chown -R quantumclaw:quantumclaw /app/logs /app/data

USER quantumclaw
EXPOSE 3666

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:3666/api/status || exit 1

ENV PORT=3666 GIN_MODE=release LOG_DIR=/app/logs

ENTRYPOINT ["dumb-init", "--"]
CMD ["./quantumclaw"]
