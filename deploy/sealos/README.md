# QuantumClaw - SealOS One-Click Deployment

## Overview
Deploy QuantumClaw on SealOS in one click using Docker Compose.

## Quick Start
1. Open the SealOS dashboard
2. Click "Create App" → "YAML Template"
3. Paste the content of `sealos.yaml`
4. Set the required environment variables
5. Click "Deploy"

## Required Environment Variables
- `SESSION_SECRET` - Random string for session encryption
- `SQL_DSN` - MySQL connection string (required in multi-node mode)
- `REDIS_CONN_STRING` - Redis connection string
- At least one model API key
