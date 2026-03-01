.PHONY: dev dev-backend dev-frontend build test lint docker deploy backup

# Local development
dev: dev-backend dev-frontend

dev-backend:
	cd backend && go run ./cmd/server

dev-worker:
	cd backend && go run ./cmd/worker

dev-frontend:
	cd frontend && npm run dev

# Build
build: build-backend build-frontend

build-backend:
	cd backend && go build -o bin/server ./cmd/server && go build -o bin/worker ./cmd/worker

build-frontend:
	cd frontend && npm run build

# Test
test: test-backend test-frontend

test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && npm test -- --run 2>/dev/null || true

# Lint
lint: lint-backend lint-frontend

lint-backend:
	cd backend && go vet ./...

lint-frontend:
	cd frontend && npm run lint 2>/dev/null || true

# Docker
docker:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

# Deploy
deploy:
	docker compose build
	docker compose up -d --no-deps frontend server worker

# Backup
backup:
	./scripts/backup.sh
