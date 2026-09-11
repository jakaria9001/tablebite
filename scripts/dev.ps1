# 1. Verify PostgreSQL
pg_isready

# 2. Apply migrations
Push-Location backend
# go run ./cmd/migrate

# 3. Seed/update menu
# go run ./cmd/seed-menu

# 4. Start backend
Start-Process powershell -ArgumentList "-NoExit", "-Command", "go run ./cmd/server"

Pop-Location

# 5. Start frontend
Push-Location frontend
npm run dev
Pop-Location