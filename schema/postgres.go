package schema

import (
	"context"
	"database/sql"
	"fmt"
)

func introspectPostgres(ctx context.Context, db *sql.DB) (*Schema, error) {
	// 1. Get Tables and Views
	rows, err := db.QueryContext(ctx, `
		SELECT
			c.table_name,
			c.column_name,
			c.data_type,
			c.is_nullable = 'YES' AS nullable,
			EXISTS (
				SELECT 1 FROM information_schema.key_column_usage kcu
				JOIN information_schema.table_constraints tc ON kcu.constraint_name = tc.constraint_name
				WHERE kcu.table_name = c.table_name AND kcu.column_name = c.column_name AND tc.constraint_type = 'PRIMARY KEY'
			) AS is_pk,
			t.table_type
		FROM information_schema.columns c
		JOIN information_schema.tables t ON c.table_name = t.table_name AND c.table_schema = t.table_schema
		WHERE c.table_schema = 'public'
		ORDER BY c.table_name, c.ordinal_position
	`)
	if err != nil {
		return nil, fmt.Errorf("introspect query: %w", err)
	}
	defer rows.Close()

	schema := &Schema{Dialect: "postgres"}
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

	// 2. Get Functions/Procedures
	// This query gets public functions that aren't internal to Postgres
	procRows, err := db.QueryContext(ctx, `
		SELECT
			p.proname AS name,
			pg_get_function_arguments(p.oid) AS args,
			pg_get_function_result(p.oid) AS return_type
		FROM pg_proc p
		JOIN pg_namespace n ON p.pronamespace = n.oid
		WHERE n.nspname = 'public'
		AND p.prokind IN ('f', 'p')
	`)
	if err != nil {
		return nil, fmt.Errorf("introspect procs: %w", err)
	}
	defer procRows.Close()

	for procRows.Next() {
		var p Procedure
		if err := procRows.Scan(&p.Name, &p.Args, &p.ReturnType); err != nil {
			return nil, fmt.Errorf("scan proc: %w", err)
		}
		schema.Procedures = append(schema.Procedures, p)
	}

	return schema, rows.Err()
}
