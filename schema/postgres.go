package schema

import (
	"context"
	"database/sql"
	"fmt"
)

func introspectPostgres(ctx context.Context, db *sql.DB) (*Schema, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			c.table_name,
			c.column_name,
			c.data_type,
			c.is_nullable = 'YES' AS nullable,
			EXISTS (
				SELECT 1
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage ku
					ON tc.constraint_name = ku.constraint_name
				WHERE tc.constraint_type = 'PRIMARY KEY'
				  AND ku.table_name  = c.table_name
				  AND ku.column_name = c.column_name
				  AND ku.table_schema = c.table_schema
			) AS is_pk
		FROM information_schema.columns c
		WHERE c.table_schema = 'public'
		ORDER BY c.table_name, c.ordinal_position
	`)
	if err != nil {
		return nil, fmt.Errorf("introspect query: %w", err)
	}
	defer rows.Close()

	schema := &Schema{Dialect: "postgres"}
	tableIndex := map[string]int{} // tableName → index in schema.Tables

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

		// Find or create the Table entry
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
