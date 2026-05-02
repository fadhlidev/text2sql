package text2sql

import (
	"context"
	"strings"
	"testing"

	"github.com/fadhlidev/text2sql/schema"
)

// mockLLM is a fake LLM that always returns the response you configure.
// It satisfies the LLMClient interface.
type mockLLM struct {
	response string
	err      error
}

func (m *mockLLM) Complete(_ context.Context, _, _ string) (string, error) {
	return m.response, m.err
}

func newTestConverter(llmResponse string) *Converter {
	s := &schema.Schema{
		Dialect: "postgres",
		Tables: []schema.Table{
			{Name: "customers", Columns: []schema.Column{
				{Name: "id", Type: "integer"},
				{Name: "name", Type: "varchar"},
			}},
		},
	}
	return New(&mockLLM{response: llmResponse}, s)
}

func TestConverter_ReturnsSQL_WhenLLMRespondsWithSelect(t *testing.T) {
	conv := newTestConverter("SELECT id, name FROM customers LIMIT 100")

	sql, err := conv.TextToSQL(context.Background(), "list all customers")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(sql, "SELECT") {
		t.Errorf("expected SQL to start with SELECT, got: %s", sql)
	}
}

func TestConverter_ReturnsError_WhenLLMReturnsErrorPrefix(t *testing.T) {
	conv := newTestConverter("ERROR: no table named invoices in schema")

	_, err := conv.TextToSQL(context.Background(), "show all invoices")

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "unanswerable") {
		t.Errorf("expected 'unanswerable' in error message, got: %v", err)
	}
}

func TestConverter_ReturnsError_WhenLLMReturnsUnsafeSQL(t *testing.T) {
	conv := newTestConverter("DROP TABLE customers")

	_, err := conv.TextToSQL(context.Background(), "delete all customers")

	if err == nil {
		t.Fatal("expected a validation error, got nil")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected 'validation failed' in error message, got: %v", err)
	}
}

func TestConverter_TrimsWhitespace(t *testing.T) {
	conv := newTestConverter("  \n  SELECT id FROM customers  \n  ")

	sql, err := conv.TextToSQL(context.Background(), "get customer ids")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.HasPrefix(sql, " ") || strings.HasSuffix(sql, " ") {
		t.Errorf("expected trimmed SQL, got: %q", sql)
	}
}
