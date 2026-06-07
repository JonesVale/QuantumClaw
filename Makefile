.PHONY: build build-fe build-go build-app build-electron build-sdk build-cmd-sync build-all clean dev sdk-release

# Build everything (frontend + Go binary)
build: build-fe build-go

# Build frontend only
build-fe:
	cd web/default && npm run build

# Build Go binary only
build-go:
	go build -o quantumclaw_new.exe .

# Build mobile app (Expo)
build-app:
	cd app && npm install && npx expo export --platform web

# Build desktop app (Electron)
build-electron:
	cd electron && npm install && npx electron-builder --win --x64

# Build cmd_sync (级联数据同步工具)
build-cmd-sync:
	go build -o cmd_sync.exe ./cmd_sync

# SDK targets
build-sdk-node:
	cd sdk/nodejs && npm install && npm run build

build-sdk-python:
	cd sdk/python && pip install build && python -m build

build-sdk: build-sdk-node build-sdk-python

# Build all (everything)
build-all: build-fe build-go build-app build-electron build-sdk build-cmd-sync

# Create SDK release packages
sdk-release:
	@echo "=== Node.js SDK ==="
	cd sdk/nodejs && npm pack
	@echo ""
	@echo "=== Python SDK ==="
	cd sdk/python && python setup.py sdist bdist_wheel
	@echo ""
	@echo "SDK packages built. See sdk/nodejs/ and sdk/python/dist/"

# Clean build artifacts
clean:
	rm -f quantumclaw_new.exe cmd_sync.exe
	cd web/default && rm -rf dist
	cd sdk/nodejs && rm -rf dist *.tgz
	cd sdk/python && rm -rf dist build *.egg-info

# Dev: frontend + Go binary + start
dev: build
	./quantumclaw_new.exe
