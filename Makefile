.PHONY: build build-fe build-go clean dev

# Build everything (frontend + Go binary)
build: build-fe build-go

# Build frontend only
build-fe:
	cd web/default && npm run build

# Build Go binary only
build-go:
	go build -o quantumclaw_new.exe .

# Clean build artifacts
clean:
	rm -f quantumclaw_new.exe
	cd web/default && rm -rf dist

# Dev: frontend + Go binary + start
dev: build
	./quantumclaw_new.exe
