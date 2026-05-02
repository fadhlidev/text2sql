package text2sql

import (
	"context"
	"strings"
	"testing"

	"github.com/fadhlidev/text2sql/schema"
)

// mockLLM is a fake LLM that always returns the response you configure.
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

func TestConverter_ReturnsSQL_WhenLLMRespondsWithJSON(t *testing.T) {
	jsonResp := `{"sql": "SELECT id, name FROM customers LIMIT 100", "explanation": "Fetching customers."}`
	conv := newTestConverter(jsonResp)

	sql, expl, err := conv.TextToSQL(context.Background(), "list all customers")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(sql, "SELECT") {
		t.Errorf("expected SQL to start with SELECT, got: %s", sql)
	}
	if expl != "Fetching customers." {
		t.Errorf("expected explanation, got: %s", expl)
	}
}

func TestConverter_ReturnsError_WhenLLMReturnsErrorPrefix(t *testing.T) {
	conv := newTestConverter("ERROR: no table named invoices in schema")

	_, _, err := conv.TextToSQL(context.Background(), "show all invoices")

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !strings.Contains(err.Error(), "unanswerable") {
		t.Errorf("expected 'unanswerable' in error message, got: %v", err)
	}
}

func TestConverter_ReturnsError_WhenLLMReturnsUnsafeSQL(t *testing.T) {
	jsonResp := `{"sql": "DROP TABLE customers", "explanation": "Deleting them all."}`
	conv := newTestConverter(jsonResp)

	_, _, err := conv.TextToSQL(context.Background(), "delete all customers")

	if err == nil {
		t.Fatal("expected a validation error, got nil")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected 'validation failed' in error message, got: %v", err)
	}
}

func TestConverter_ReturnsError_WhenInvalidJSON(t *testing.T) {
	conv := newTestConverter("not a json")

	_, _, err := conv.TextToSQL(context.Background(), "test")

	if err == nil {
		t.Fatal("expected error for invalid json, got nil")
	}
}
