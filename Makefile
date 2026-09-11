.PHONY: dev migrate seed backend frontend backend-test frontend-check test

dev:
	./scripts/dev.sh

migrate:
	cd backend && go run ./cmd/migrate

seed:
	cd backend && go run ./cmd/seed-menu

backend:
	cd backend && go run ./cmd/server

frontend:
	cd frontend && npm run dev

backend-test:
	cd backend && go test ./...

frontend-check:
	cd frontend && npm run build && npm run lint

test: backend-test frontend-check
