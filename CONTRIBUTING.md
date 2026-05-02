# Contributing to Text2SQL

Thank you for your interest in contributing to Text2SQL! We welcome all contributions, from bug reports and documentation to new features and provider implementations.

## Development Setup

### Prerequisites
- [Nix](https://nixos.org/download.html) with Flakes enabled.
- Docker (for database and testing).

### Getting Started
1. Clone the repository.
2. Run `nix develop` to enter the development environment.
3. Start the development database:
   ```bash
   docker compose up -d
   ```
4. Copy `.env.example` to `.env` and fill in your API keys.

## Running Tests

We prioritize high test coverage. Please ensure all tests pass before submitting a PR.

### Unit Tests
```bash
go test ./...
```

### Integration Tests
Integration tests require a running test database:
```bash
docker compose -f docker-compose.test.yml up -d
export TEST_DB_URI="postgres://testuser:testpassword@localhost:5433/text2sql_test"
go test ./... -v
```

## Branching & Pull Requests
- Base all new features on the `main` branch.
- Use descriptive branch names like `feat/new-llm-provider` or `fix/schema-detection`.
- Follow the commit message format: `[action]: [message]` (e.g., `add: support for Mistral`, `fix: handle null columns in MySQL`).

## Code Style
- Run `go fmt ./...` before committing.
- Ensure all new logic includes unit tests.
- Keep the `README.md` and documentation up to date.

## License
By contributing, you agree that your contributions will be licensed under the MIT License.
