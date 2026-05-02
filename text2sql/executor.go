package text2sql

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Executor runs SQL queries against the database and returns rows as JSON-friendly maps.
type Executor struct {
	db *sql.DB
}

func NewExecutor(db *sql.DB) *Executor {
	return &Executor{db: db}
}

// Run executes a SQL query and returns the results as a slice of maps.
// Each map represents one row: the key is the column name, the value is the cell value.
// Always returns an empty slice (not nil) when there are no results.
func (e *Executor) Run(ctx context.Context, query string) ([]map[string]any, error) {
	// Apply a hard timeout so slow queries don't block the server
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	// Get column names from the result set
	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("columns: %w", err)
	}

	var result []map[string]any

	for rows.Next() {
		// Allocate one interface{} pointer per column to receive scanned values
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		// Build the row map
		row := make(map[string]any, len(cols))
		for i, col := range cols {
			// Convert []byte to string — otherwise JSON output will be base64-encoded
			if b, ok := vals[i].([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = vals[i]
			}
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	// Always return an array, never nil (nil serializes to JSON null)
	if result == nil {
		result = []map[string]any{}
	}

	return result, nil
}
