package testhelper

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// OpenTestDB connects to the test database using TEST_DB_URI.
// It infers the driver from the URI.
func OpenTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	uri := os.Getenv("TEST_DB_URI")
	if uri == "" {
		t.Skip("TEST_DB_URI not set — skipping integration test")
	}

	driver := "pgx"
	if strings.HasPrefix(uri, "mysql") {
		driver = "mysql"
	} else if strings.HasPrefix(uri, "sqlite") || !strings.Contains(uri, "://") {
		driver = "sqlite"
	}

	db, err := sql.Open(driver, uri)
	if err != nil {
		t.Fatalf("testhelper: open db (%s): %v", driver, err)
	}

	// SQLite doesn't always support pinging if the file doesn't exist yet,
	// but for our tests it should be fine.
	if driver != "sqlite" {
		if err := db.PingContext(context.Background()); err != nil {
			t.Fatalf("testhelper: ping db: %v", err)
		}
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

// Seed creates a minimal schema and inserts test rows.
// The SQL is slightly adjusted for dialect compatibility (SERIAL vs AUTO_INCREMENT).
func Seed(t *testing.T, db *sql.DB, dialect string) {
	t.Helper()

	ctx := context.Background()

	var queries []string
	switch dialect {
	case "postgres":
		queries = []string{
			"DROP TABLE IF EXISTS orders CASCADE",
			"DROP TABLE IF EXISTS customers CASCADE",
			`CREATE TABLE customers (
				id         SERIAL PRIMARY KEY,
				name       VARCHAR(100) NOT NULL,
				email      VARCHAR(100) NOT NULL
			)`,
			`CREATE TABLE orders (
				id          SERIAL PRIMARY KEY,
				customer_id INTEGER NOT NULL REFERENCES customers(id),
				total       NUMERIC(10, 2) NOT NULL,
				created_at  TIMESTAMP DEFAULT NOW()
			)`,
		}
	case "mysql":
		queries = []string{
			"SET FOREIGN_KEY_CHECKS = 0",
			"DROP TABLE IF EXISTS orders",
			"DROP TABLE IF EXISTS customers",
			"SET FOREIGN_KEY_CHECKS = 1",
			`CREATE TABLE customers (
				id         INT AUTO_INCREMENT PRIMARY KEY,
				name       VARCHAR(100) NOT NULL,
				email      VARCHAR(100) NOT NULL
			)`,
			`CREATE TABLE orders (
				id          INT AUTO_INCREMENT PRIMARY KEY,
				customer_id INT NOT NULL,
				total       DECIMAL(10, 2) NOT NULL,
				created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (customer_id) REFERENCES customers(id)
			)`,
		}
	case "sqlite":
		queries = []string{
			"DROP TABLE IF EXISTS orders",
			"DROP TABLE IF EXISTS customers",
			`CREATE TABLE customers (
				id         INTEGER PRIMARY KEY AUTOINCREMENT,
				name       TEXT NOT NULL,
				email      TEXT NOT NULL
			)`,
			`CREATE TABLE orders (
				id          INTEGER PRIMARY KEY AUTOINCREMENT,
				customer_id INTEGER NOT NULL REFERENCES customers(id),
				total       DECIMAL(10, 2) NOT NULL,
				created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)`,
		}
	}

	for _, q := range queries {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatalf("testhelper: seed (ddl): %v", err)
		}
	}

	_, err := db.ExecContext(ctx, `
        INSERT INTO customers (name, email) VALUES
            ('Alice', 'alice@example.com'),
            ('Bob',   'bob@example.com');
    `)
	if err != nil {
		t.Fatalf("testhelper: seed (insert customers): %v", err)
	}

	_, err = db.ExecContext(ctx, `
        INSERT INTO orders (customer_id, total) VALUES
            (1, 100.00),
            (1, 250.50),
            (2, 75.00);
    `)
	if err != nil {
		t.Fatalf("testhelper: seed (insert orders): %v", err)
	}
}
