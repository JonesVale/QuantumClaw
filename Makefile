.PHONY: build build-fe build-go build-app build-electron clean dev

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

# Build all (everything including app/electron)
build-all: build-fe build-go build-app build-electron

# Clean build artifacts
clean:
	rm -f quantumclaw_new.exe
	cd web/default && rm -rf dist

# Dev: frontend + Go binary + start
dev: build
	./quantumclaw_new.exe
