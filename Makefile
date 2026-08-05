.PHONY: dev build clean

frontend-install:
	cd frontend && npm install

dev:
	wails dev

build:
	wails build -platform windows/amd64 -o netsight.exe

clean:
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	go clean

test:
	go test ./internal/...

fmt:
	go fmt ./...
	cd frontend && npx prettier --write "src/**/*.{ts,tsx,css}"
