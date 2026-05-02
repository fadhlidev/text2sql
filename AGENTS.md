# AGENTS.md

## Commands

```bash
# Dev: start dev DB, start app
docker compose up -d
go run main.go

# Integration tests: start test DB, source env, run tests
docker compose -f docker-compose.test.yml up -d
# Set TEST_DB_URI (e.g., from .env.test or manually)
export TEST_DB_URI="postgres://testuser:testpassword@localhost:5433/text2sql_test"
go test ./...
```

- **Dev DB**: port 5432, container `text2sql_dev_db`
- **Test DB**: port 5433, container `text2sql_test_db` (tmpfs for speed)
- Skip integration tests if `TEST_DB_URI` is not set

## Architecture

- Schema loaded once at startup, cached in memory (not per-request)
- LLM temperature is always 0
- SQL execution has 10-second timeout
- DB user must be read-only

## Environment

Required in `.env`:
- `DB_URI` — PostgreSQL connection string
- `OPENAI_API_KEY` — OpenAI API key

`TEST_DB_URI` must be set (with port 5433) for integration tests. `.env.test` can be committed.

## Commit Messages
Format: `[action]: [message]` (e.g., `add: new endpoint`, `fix: memory leak in cache`)