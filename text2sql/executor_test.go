package text2sql

import (
	"context"
	"testing"

	"github.com/fadhlidev/text2sql/testhelper"
)

func TestExecutor_ReturnsRows(t *testing.T) {
	db, cleanup := testhelper.OpenTestDB(t)
	defer cleanup()
	testhelper.Seed(t, db, "postgres")

	exec := NewExecutor(db)
	result, err := exec.Run(context.Background(), "SELECT id, name FROM customers ORDER BY id")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 rows, got %d", len(result))
	}
	if result[0]["name"] != "Alice" {
		t.Errorf("expected first row name to be 'Alice', got: %v", result[0]["name"])
	}
}

func TestExecutor_ReturnsEmptySlice_WhenNoRows(t *testing.T) {
	db, cleanup := testhelper.OpenTestDB(t)
	defer cleanup()
	testhelper.Seed(t, db, "postgres")

	exec := NewExecutor(db)
	result, err := exec.Run(context.Background(), "SELECT id FROM customers WHERE id = 99999")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Must be an empty slice, never nil — nil serializes to JSON null
	if result == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(result) != 0 {
		t.Errorf("expected 0 rows, got %d", len(result))
	}
}

func TestExecutor_RespectsTimeout(t *testing.T) {
	db, cleanup := testhelper.OpenTestDB(t)
	defer cleanup()

	exec := NewExecutor(db)

	// pg_sleep(30) will be cancelled by the executor's 10-second timeout
	_, err := exec.Run(context.Background(), "SELECT pg_sleep(30)")

	if err == nil {
		t.Error("expected a timeout error, got nil")
	}
}

func TestExecutor_ReturnsAggregates(t *testing.T) {
	db, cleanup := testhelper.OpenTestDB(t)
	defer cleanup()
	testhelper.Seed(t, db, "postgres")

	exec := NewExecutor(db)
	result, err := exec.Run(context.Background(),
		"SELECT customer_id, SUM(total) AS total FROM orders GROUP BY customer_id ORDER BY customer_id")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 groups, got %d", len(result))
	}
}
