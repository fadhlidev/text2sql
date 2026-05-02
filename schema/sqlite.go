package schema

import (
	"context"
	"database/sql"
	"fmt"
)

func introspectSQLite(ctx context.Context, db *sql.DB) (*Schema, error) {
	// 1. Get all table and view names
	tableRows, err := db.QueryContext(ctx, "SELECT name, type FROM sqlite_master WHERE type IN ('table', 'view') AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer tableRows.Close()

	schema := &Schema{Dialect: "sqlite"}

	for tableRows.Next() {
		var tableName string
		var tableType string
		if err := tableRows.Scan(&tableName, &tableType); err != nil {
			return nil, fmt.Errorf("scan table name: %w", err)
		}

		table := Table{Name: tableName, Type: tableType}

		// 2. Get columns for this table/view
		colRows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
		if err != nil {
			return nil, fmt.Errorf("table info for %s: %w", tableName, err)
		}

		for colRows.Next() {
			var (
				cid       int
				name      string
				dataType  string
				notnull   int
				dfltValue any
				pk        int
			)
			if err := colRows.Scan(&cid, &name, &dataType, &notnull, &dfltValue, &pk); err != nil {
				colRows.Close()
				return nil, fmt.Errorf("scan column info: %w", err)
			}

			table.Columns = append(table.Columns, Column{
				Name:     name,
				Type:     dataType,
				Nullable: notnull == 0,
				IsPK:     pk > 0,
			})
		}
		colRows.Close()
		schema.Tables = append(schema.Tables, table)
	}

	return schema, tableRows.Err()
}
