# TableBite

A restaurant website and dine-in QR ordering platform built with the same core stack as ShuttleHub: React + TypeScript + Vite + Tailwind CSS on the frontend, and Go + Chi + PostgreSQL + pgx on the backend.

## What is in this skeleton

- React 19 + TypeScript + Vite frontend
- Tailwind CSS v4 via the Vite plugin
- React Router
- Go HTTP API using Chi
- PostgreSQL via pgxpool
- SQL migration runner implemented in Go
- JSON menu importer
- Table/QR data model ready; QR image generation is intentionally deferred to the table-management slice
- Seeded menu schema supporting single-price and half/full pricing
- Basic public menu API
- Basic admin-auth command placeholder
- `.env.example`
- Windows PowerShell and Unix shell dev scripts

## Menu seed

The provided `backend/seed/menu.json` is copied from the supplied source file and currently contains 25 categories and 175 menu items. It is treated as an import/seed source, not as the production data store.

The importer preserves:
- category order
- menu item order
- descriptions
- INR currency
- veg/non-veg flag
- availability
- single price
- half/full pricing
- source IDs
- image URLs

## Prerequisites

- Node.js 20.19+ or 22.12+
- npm
- Go (the project targets the ShuttleHub-era Go stack; use Go 1.27+ for the intended environment)
- PostgreSQL 14+

## Local setup

1. Copy `.env.example` to `.env` and adjust PostgreSQL credentials if needed.
2. Ensure PostgreSQL is running and the `tablebite` database exists.
3. From PowerShell:

```powershell
.\scripts\dev.ps1
```

From Bash:

```bash
./scripts/dev.sh
```

Those scripts install frontend dependencies when needed, download Go dependencies, run migrations, seed the menu, and start the Go API + Vite frontend.

## Manual commands

Backend:

```bash
cd backend
go mod download
go run ./cmd/migrate
go run ./cmd/seed-menu
go run ./cmd/server

Create/update the first admin:

```bash
ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD="change-me" go run ./cmd/create-admin
```
```

Frontend:

```bash
cd frontend
npm install
npm run dev
```

## API

- `GET /api/health`
- `GET /api/v1/menu`
- `GET /api/v1/categories`

## Next implementation slice

The skeleton is intentionally focused on reliable foundations. The next slice should implement:

1. Home page sections
2. Full menu filtering/search
3. Table + QR token management
4. Cart
5. Table session
6. Order creation/status lifecycle
7. Admin login/RBAC
8. Incoming-order dashboard
9. WhatsApp/SMS notification adapters
