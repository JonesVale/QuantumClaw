# ============================================================
# Stage 1: Build Frontend
# ============================================================
FROM node:22-alpine AS frontend-builder

WORKDIR /app

# Copy only package files first (for better Docker layer caching)
COPY web/default/package*.json ./default/

# Install dependencies
RUN cd default && npm install && cd ..

# Copy frontend source files
COPY web/default ./default/

# Build frontend with Rsbuild, output to /app/web/build/default
ARG VERSION=dev
RUN mkdir -p /app/web/build && \
    cd default && \
    VITE_REACT_APP_VERSION=${VERSION} npx rsbuild build && \
    cd .. && \
    mv default/dist /app/web/build/default && \
    echo "Build output:" && ls -la /app/web/build/default/

# ============================================================
# Stage 2: Build Go Backend
# ============================================================
FROM golang:1.23-alpine AS backend-builder

# 瀹夎 CGO 渚濊禆锛圫QLite3 闇€瑕侊級
RUN apk add --no-cache git ca-certificates tzdata build-base

WORKDIR /app

# 璁剧疆鍥藉唴 Go 妯″潡浠ｇ悊锛堣В鍐冲浗鍐呯綉缁滈棶棰橈級
ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,direct
ENV GOSUMDB=off

# Install goxz for cross-compilation (optional, for smaller image)
# COPY --from=frontend-builder /app/web/build ./web/build

# Copy go mod files first
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Copy built frontend
COPY --from=frontend-builder /app/web/build/default ./web/default/dist

# Copy static assets to dist directory
COPY web/default/public/logo.webp ./web/default/dist/logo.webp

COPY web/default/public/favicon.ico ./web/default/dist/favicon.ico

# Build with version info (CGO_ENABLED=1 for SQLite3 driver)
ARG VERSION=dev
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -X github.com/quantumclaw/quantumclaw/common.Version=${VERSION}" \
    -o quantumclaw .

# ============================================================
# Stage 3: Final Runtime Image
# ============================================================
FROM alpine:3.19

LABEL maintainer="QuantumClaw <quantumclaw@example.com>"
LABEL description="QuantumClaw - AI API Gateway & Token Distribution Platform"

# Install runtime dependencies
RUN apk add --no-cache \
        ca-certificates \
        tzdata \
        dumb-init \
        su-exec \
    && addgroup -g 1000 quantumclaw \
    && adduser -u 1000 -G quantumclaw -s /bin/sh -D quantumclaw

WORKDIR /app

# Copy binary from builder
COPY --from=backend-builder /app/quantumclaw .

# Copy static files (if any)
COPY --from=backend-builder /app/web ./web

# Create data and logs directories
RUN mkdir -p /app/logs /app/data && chown -R quantumclaw:quantumclaw /app/logs /app/data

USER quantumclaw

# Expose default port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:3000/api/status || exit 1

# Default environment variables
ENV PORT=3000
ENV GIN_MODE=release
ENV LOG_DIR=/app/logs

ENTRYPOINT ["dumb-init", "--"]
CMD ["./quantumclaw"]

