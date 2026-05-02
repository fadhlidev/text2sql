package text2sql

import "testing"

func TestValidate_AllowsSelect(t *testing.T) {
	cases := []string{
		"SELECT id FROM customers",
		"SELECT id FROM customers LIMIT 10",
		"  SELECT * FROM orders WHERE id = 1", // leading whitespace is fine
		"WITH cte AS (SELECT 1) SELECT * FROM cte",
	}
	for _, sql := range cases {
		if err := validate(sql); err != nil {
			t.Errorf("expected %q to pass, got error: %v", sql, err)
		}
	}
}

func TestValidate_BlocksNonSelect(t *testing.T) {
	cases := []struct {
		sql  string
		desc string
	}{
		{"INSERT INTO customers VALUES (1)", "INSERT"},
		{"UPDATE customers SET name='x' WHERE id=1", "UPDATE"},
		{"DELETE FROM customers", "DELETE"},
		{"DROP TABLE customers", "DROP"},
		{"ALTER TABLE customers ADD COLUMN age INT", "ALTER"},
		{"TRUNCATE TABLE customers", "TRUNCATE"},
		{"CREATE TABLE foo (id INT)", "CREATE"},
		{"GRANT SELECT ON customers TO public", "GRANT"},
		{"SELECT 1; DROP TABLE customers", "inline DROP after SELECT"},
		{"SELECT 1 -- this is a comment", "SQL comment injection"},
		{"SELECT 1 /* block comment */", "block comment injection"},
	}
	for _, tc := range cases {
		if err := validate(tc.sql); err == nil {
			t.Errorf("expected %q (%s) to be blocked, but it passed", tc.sql, tc.desc)
		}
	}
}
