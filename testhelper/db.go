package testhelper

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// OpenTestDB connects to the test database using TEST_DB_URI.
// If TEST_DB_URI is not set, the test is skipped (not failed).
// Always call the returned cleanup function at the end of your test:
//
//   db, cleanup := testhelper.OpenTestDB(t)
//   defer cleanup()
func OpenTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper() // marks this function as a helper so test line numbers are accurate

	uri := os.Getenv("TEST_DB_URI")
	if uri == "" {
		t.Skip("TEST_DB_URI not set — skipping integration test")
	}

	db, err := sql.Open("pgx", uri)
	if err != nil {
		t.Fatalf("testhelper: open db: %v", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("testhelper: ping db: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

// Seed creates a minimal schema and inserts test rows for use in executor and introspection tests.
// It drops and recreates the tables every time so each test starts with a clean state.
func Seed(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx := context.Background()

	_, err := db.ExecContext(ctx, `
        DROP TABLE IF EXISTS orders;
        DROP TABLE IF EXISTS customers;

        CREATE TABLE customers (
            id         SERIAL PRIMARY KEY,
            name       VARCHAR(100) NOT NULL,
            email      VARCHAR(100) NOT NULL
        );

        CREATE TABLE orders (
            id          SERIAL PRIMARY KEY,
            customer_id INTEGER NOT NULL REFERENCES customers(id),
            total       NUMERIC(10, 2) NOT NULL,
            created_at  TIMESTAMP DEFAULT NOW()
        );

        INSERT INTO customers (name, email) VALUES
            ('Alice', 'alice@example.com'),
            ('Bob',   'bob@example.com');

        INSERT INTO orders (customer_id, total) VALUES
            (1, 100.00),
            (1, 250.50),
            (2, 75.00);
    `)
	if err != nil {
		t.Fatalf("testhelper: seed: %v", err)
	}
}
