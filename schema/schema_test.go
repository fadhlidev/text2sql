package schema

import (
	"context"
	"strings"
	"testing"

	"github.com/fadhlidev/text2sql/testhelper"
)

func TestSchema_Context_ContainsTableAndColumns(t *testing.T) {
	s := &Schema{
		Dialect: "postgres",
		Tables: []Table{
			{
				Name: "customers",
				Columns: []Column{
					{Name: "id", Type: "integer", IsPK: true, Nullable: false},
					{Name: "email", Type: "varchar", IsPK: false, Nullable: false},
					{Name: "phone", Type: "varchar", IsPK: false, Nullable: true},
				},
			},
		},
	}

	ctx := s.Context()

	// The output must mention the table name
	if !strings.Contains(ctx, "Table: customers") {
		t.Error("expected context to contain 'Table: customers'")
	}

	// Primary key columns must be annotated
	if !strings.Contains(ctx, "[PK]") {
		t.Error("expected context to contain '[PK]' annotation")
	}

	// NOT NULL must appear for non-nullable columns
	if !strings.Contains(ctx, "NOT NULL") {
		t.Error("expected context to contain 'NOT NULL'")
	}

	// Nullable columns must NOT have NOT NULL
	if strings.Contains(ctx, "phone varchar NOT NULL") {
		t.Error("nullable column 'phone' should not be marked NOT NULL")
	}
}

func TestSchema_Context_IncludesForeignKey(t *testing.T) {
	s := &Schema{
		Tables: []Table{
			{
				Name: "orders",
				Columns: []Column{
					{Name: "customer_id", Type: "integer", FKRef: "customers.id"},
				},
			},
		},
	}

	ctx := s.Context()

	if !strings.Contains(ctx, "FK→customers.id") {
		t.Errorf("expected FK annotation, got:\n%s", ctx)
	}
}

func TestIntrospect_Postgres_ReadsTablesAndColumns(t *testing.T) {
	db, cleanup := testhelper.OpenTestDB(t)
	defer cleanup()
	testhelper.Seed(t, db)

	s, err := Introspect(context.Background(), db, "postgres")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Tables) == 0 {
		t.Fatal("expected at least one table, got none")
	}

	// Find the customers table
	var customersTable *Table
	for i := range s.Tables {
		if s.Tables[i].Name == "customers" {
			customersTable = &s.Tables[i]
			break
		}
	}
	if customersTable == nil {
		t.Fatal("expected 'customers' table in schema, not found")
	}

	// Verify the primary key column is detected
	var foundPK bool
	for _, col := range customersTable.Columns {
		if col.Name == "id" && col.IsPK {
			foundPK = true
		}
	}
	if !foundPK {
		t.Error("expected 'id' column to be detected as primary key")
	}
}
