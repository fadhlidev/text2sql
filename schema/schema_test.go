package schema

import (
	"context"
	"os"
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
	testhelper.Seed(t, db, "postgres")

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

func TestIntrospect_MySQL_ReadsTablesAndColumns(t *testing.T) {
	// To run this, you need a MySQL instance and TEST_DB_URI set to it.
	// e.g., TEST_DB_URI=mysql://user:pass@tcp(localhost:3306)/testdb
	if !strings.Contains(os.Getenv("TEST_DB_URI"), "mysql") {
		t.Skip("TEST_DB_URI is not a MySQL URI — skipping")
	}

	db, cleanup := testhelper.OpenTestDB(t)
	defer cleanup()
	testhelper.Seed(t, db, "mysql")

	s, err := Introspect(context.Background(), db, "mysql")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Tables) == 0 {
		t.Fatal("expected at least one table, got none")
	}

	// Check for customers table
	var found bool
	for _, table := range s.Tables {
		if table.Name == "customers" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'customers' table to be introspected")
	}
}

func TestIntrospect_SQLite_ReadsTablesAndColumns(t *testing.T) {
	// We can use an in-memory SQLite database for this test
	os.Setenv("TEST_DB_URI", "file::memory:?cache=shared")
	defer os.Unsetenv("TEST_DB_URI")

	db, cleanup := testhelper.OpenTestDB(t)
	defer cleanup()
	testhelper.Seed(t, db, "sqlite")

	s, err := Introspect(context.Background(), db, "sqlite")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// SQLite introspection should find our 2 tables
	if len(s.Tables) != 2 {
		t.Errorf("expected 2 tables, got %d", len(s.Tables))
	}

	var customersTable *Table
	for i := range s.Tables {
		if s.Tables[i].Name == "customers" {
			customersTable = &s.Tables[i]
		}
	}

	if customersTable == nil {
		t.Fatal("'customers' table not found in SQLite schema")
	}

	// Verify columns
	if len(customersTable.Columns) != 3 {
		t.Errorf("expected 3 columns for customers, got %d", len(customersTable.Columns))
	}

	var foundPK bool
	for _, col := range customersTable.Columns {
		if col.Name == "id" && col.IsPK {
			foundPK = true
		}
	}
	if !foundPK {
		t.Error("expected 'id' column to be primary key in SQLite")
	}
}
