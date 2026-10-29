.PHONY: frontend backend build run dev-backend dev-frontend check test test-e2e clean

frontend:
	cd frontend && pnpm install --frozen-lockfile && pnpm run build
	rm -rf backend/cmd/server/static
	mkdir -p backend/cmd/server/static
	cp -R frontend/dist/. backend/cmd/server/static/
	touch backend/cmd/server/static/.gitkeep

backend:
	cd backend && CGO_ENABLED=0 go build -trimpath -o ../app ./cmd/server

build: frontend
	$(MAKE) backend

run: build
	./app

dev-backend:
	cd backend && go run ./cmd/server

dev-frontend:
	cd frontend && pnpm run dev

check:
	cd frontend && pnpm run check
	cd backend && go vet ./...

test:
	cd backend && go test -race ./...

test-e2e: build
	cd frontend && pnpm run test:e2e

clean:
	rm -rf frontend/dist app
	mkdir -p backend/cmd/server/static
	find backend/cmd/server/static -mindepth 1 ! -path backend/cmd/server/static/.gitkeep -delete
	touch backend/cmd/server/static/.gitkeep
