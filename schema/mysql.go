package schema

import (
	"context"
	"database/sql"
	"fmt"
)

func introspectMySQL(ctx context.Context, db *sql.DB) (*Schema, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			table_name,
			column_name,
			data_type,
			is_nullable = 'YES' AS nullable,
			column_key = 'PRI' AS is_pk
		FROM information_schema.columns
		WHERE table_schema = DATABASE()
		ORDER BY table_name, ordinal_position
	`)
	if err != nil {
		return nil, fmt.Errorf("introspect query: %w", err)
	}
	defer rows.Close()

	schema := &Schema{Dialect: "mysql"}
	tableIndex := map[string]int{}

	for rows.Next() {
		var (
			tableName  string
			columnName string
			dataType   string
			nullable   bool
			isPK       bool
		)
		if err := rows.Scan(&tableName, &columnName, &dataType, &nullable, &isPK); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		idx, exists := tableIndex[tableName]
		if !exists {
			schema.Tables = append(schema.Tables, Table{Name: tableName})
			idx = len(schema.Tables) - 1
			tableIndex[tableName] = idx
		}

		schema.Tables[idx].Columns = append(schema.Tables[idx].Columns, Column{
			Name:     columnName,
			Type:     dataType,
			Nullable: nullable,
			IsPK:     isPK,
		})
	}

	return schema, rows.Err()
}
