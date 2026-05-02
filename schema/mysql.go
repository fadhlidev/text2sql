package schema

import (
	"context"
	"database/sql"
	"fmt"
)

func introspectMySQL(ctx context.Context, db *sql.DB) (*Schema, error) {
	// 1. Tables and Views
	rows, err := db.QueryContext(ctx, `
		SELECT
			c.table_name,
			c.column_name,
			c.data_type,
			c.is_nullable = 'YES' AS nullable,
			c.column_key = 'PRI' AS is_pk,
			t.table_type
		FROM information_schema.columns c
		JOIN information_schema.tables t ON c.table_name = t.table_name AND c.table_schema = t.table_schema
		WHERE c.table_schema = DATABASE()
		ORDER BY c.table_name, c.ordinal_position
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
			tableType  string
		)
		if err := rows.Scan(&tableName, &columnName, &dataType, &nullable, &isPK, &tableType); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		idx, exists := tableIndex[tableName]
		if !exists {
			tType := "table"
			if tableType == "VIEW" {
				tType = "view"
			}
			schema.Tables = append(schema.Tables, Table{Name: tableName, Type: tType})
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

	// 2. Routines (Procedures/Functions)
	routineRows, err := db.QueryContext(ctx, `
		SELECT
			routine_name,
			'' AS args, -- MySQL doesn't have a simple function to get arg string like PG
			routine_type
		FROM information_schema.routines
		WHERE routine_schema = DATABASE()
	`)
	if err != nil {
		// Routines might require special permissions, don't fail entire introspection if it fails
		return schema, nil
	}
	defer routineRows.Close()

	for routineRows.Next() {
		var r Procedure
		if err := routineRows.Scan(&r.Name, &r.Args, &r.ReturnType); err != nil {
			continue
		}
		schema.Procedures = append(schema.Procedures, r)
	}

	return schema, rows.Err()
}
