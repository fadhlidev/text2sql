# Text2SQL

Text2SQL is a powerful, secure, and multi-dialect API that converts natural language questions into executable SQL queries. It automatically introspects your database schema and uses advanced LLMs to generate precise, read-only queries.

## Features

- **Multi-Dialect Support**: Works with PostgreSQL, MySQL, and SQLite.
- **Multi-LLM Agnostic**: Supports OpenAI, Anthropic (Claude), Google Gemini, and Azure OpenAI.
- **Security First**: 
  - Strictly enforces `SELECT` / `WITH` queries only.
  - Blocklist for dangerous keywords (DROP, DELETE, etc.).
  - Hard 10-second query timeout.
- **Automatic Introspection**: Automatically reads your tables and columns—no manual schema definition required.
- **Production Ready**: Minimal Docker image (~15MB) and Fiber-powered REST API.

## Installation

### Using Nix (Recommended)
```bash
nix develop
go run main.go
```

### Using Docker
```bash
docker build -t text2sql .
docker run -p 3000:3000 --env-file .env text2sql
```

## Configuration

The application is configured entirely via environment variables. Create a `.env` file in the root directory:

### 1. Database Configuration
| Variable | Description | Example |
|----------|-------------|---------|
| `DB_URI` | Connection string | `postgres://user:pass@localhost:5432/db` |

*Dialect is automatically inferred from the URI prefix (`postgres`, `mysql`, or `sqlite`).*

### 2. LLM Configuration
| Variable | Description | Options |
|----------|-------------|---------|
| `LLM_PROVIDER` | The AI provider to use | `openai`, `anthropic`, `gemini`, `azure-openai` |
| `LLM_MODEL` | Specific model ID | `gpt-4o`, `claude-3-5-sonnet-20240620`, etc. |

#### Provider-Specific Keys
- **OpenAI**: `OPENAI_API_KEY`
- **Anthropic**: `ANTHROPIC_API_KEY`
- **Gemini**: `GEMINI_API_KEY`
- **Azure OpenAI**: `AZURE_OPENAI_API_KEY`, `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_DEPLOYMENT_ID`

## API Usage

### `POST /query`
Convert a question to SQL and execute it.

**Request Body:**
```json
{
  "question": "how many customers joined last month?"
}
```

**Response (Success):**
```json
{
  "sql": "SELECT COUNT(*) FROM customers WHERE created_at > '2026-04-01' LIMIT 100",
  "result": [
    { "count": 142 }
  ]
}
```

**Response (Error):**
```json
{
  "error": "unanswerable: no column 'joined_at' found in table 'customers'",
  "stage": "generate"
}
```

### `GET /health`
Returns `{"status": "ok"}`.

## Development

See [CONTRIBUTING.md](./CONTRIBUTING.md) for details on setting up the development environment and running tests.

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
