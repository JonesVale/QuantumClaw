#!/bin/sh
# Build Go binary for QuantumClaw with CGO enabled
apk add --no-cache git ca-certificates tzdata build-base
cd /app
go mod download
go build -ldflags="-s -w -X github.com/quantumclaw/quantumclaw/common.Version=dev" -o quantumclaw .
