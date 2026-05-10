# ============================================================
# Stage 1: Build Frontend
# ============================================================
FROM node:20-alpine AS frontend-builder

WORKDIR /app

# Copy only package files first (for better Docker layer caching)
COPY web/package*.json web/THEMES ./
COPY web/default/package*.json ./default/
COPY web/berry/package*.json ./berry/
COPY web/air/package*.json ./air/

# Install dependencies for all themes
RUN echo "default" > THEMES && \
    cd default && npm install --legacy-peer-deps && \
    cd ../berry && npm install --legacy-peer-deps && \
    cd ../air && npm install --legacy-peer-deps && \
    cd ..

# Copy theme source files
COPY web/default ./default/
COPY web/berry ./berry/
COPY web/air ./air/

# Build all themes (output goes to web/build/<theme>)
ARG VERSION=dev
RUN for theme in default berry air; do \
        echo "Building theme: $theme" && \
        cd "$theme" && \
        REACT_APP_VERSION=${VERSION} DISABLE_ESLINT_PLUGIN=true npm run build && \
        cd .. && \
        mv "$theme/build" web/build/"$theme" || true; \
    done && \
    mkdir -p web/build && \
    ls -la web/build/

# ============================================================
# Stage 2: Build Go Backend
# ============================================================
FROM golang:1.22-alpine AS backend-builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Install goxz for cross-compilation (optional, for smaller image)
# COPY --from=frontend-builder /app/web/build ./web/build

# Copy go mod files first
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Copy built frontend
COPY --from=frontend-builder /app/web/build ./web/build

# Build with version info
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
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

# Create logs directory
RUN mkdir -p /app/logs && chown quantumclaw:quantumclaw /app/logs

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
