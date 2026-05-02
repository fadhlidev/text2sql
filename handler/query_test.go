package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// mockConverter satisfies the converter interface for handler tests
type mockConverter struct {
	sql  string
	expl string
	err  error
}

func (m *mockConverter) TextToSQL(_ context.Context, _ string) (string, string, error) {
	return m.sql, m.expl, m.err
}

// mockExecutor satisfies the executor interface for handler tests
type mockExecutor struct {
	result []map[string]any
	err    error
}

func (m *mockExecutor) Run(_ context.Context, _ string) ([]map[string]any, error) {
	return m.result, m.err
}

// newTestApp builds a minimal Fiber app with the query handler wired up
func newTestApp(conv converterIface, exec executorIface) *fiber.App {
	app := fiber.New()
	qh := &QueryHandler{conv: conv, exec: exec}
	app.Post("/query", qh.Query)
	return app
}

func TestQueryHandler_Success(t *testing.T) {
	app := newTestApp(
		&mockConverter{sql: "SELECT id FROM customers LIMIT 100", expl: "test explanation"},
		&mockExecutor{result: []map[string]any{{"id": 1}}},
	)

	body, _ := json.Marshal(map[string]string{"question": "list customers"})
	req := httptest.NewRequest("POST", "/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)

	if result["sql"] != "SELECT id FROM customers LIMIT 100" {
		t.Errorf("unexpected sql field: %v", result["sql"])
	}
	if result["explanation"] != "test explanation" {
		t.Errorf("unexpected explanation field: %v", result["explanation"])
	}
	if result["result"] == nil {
		t.Error("expected result field to be present")
	}
}

func TestQueryHandler_EmptyQuestion_Returns400(t *testing.T) {
	app := newTestApp(&mockConverter{}, &mockExecutor{})

	body, _ := json.Marshal(map[string]string{"question": "  "})
	req := httptest.NewRequest("POST", "/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	if resp.StatusCode != 400 {
		t.Errorf("expected 400 for empty question, got %d", resp.StatusCode)
	}
}

func TestQueryHandler_LLMError_Returns422(t *testing.T) {
	app := newTestApp(
		&mockConverter{err: fmt.Errorf("unanswerable: no table 'invoices'")},
		&mockExecutor{},
	)

	body, _ := json.Marshal(map[string]string{"question": "show invoices"})
	req := httptest.NewRequest("POST", "/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	if resp.StatusCode != 422 {
		t.Errorf("expected 422, got %d", resp.StatusCode)
	}

	var errResp map[string]string
	json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp["stage"] != "generate" {
		t.Errorf("expected stage='generate', got %q", errResp["stage"])
	}
}

func TestQueryHandler_ExecutorError_Returns500(t *testing.T) {
	app := newTestApp(
		&mockConverter{sql: "SELECT id FROM customers"},
		&mockExecutor{err: fmt.Errorf("query: column does not exist")},
	)

	body, _ := json.Marshal(map[string]string{"question": "list customers"})
	req := httptest.NewRequest("POST", "/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	if resp.StatusCode != 500 {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}

	var errResp map[string]string
	json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp["stage"] != "execute" {
		t.Errorf("expected stage='execute', got %q", errResp["stage"])
	}
}

func TestQueryHandler_InvalidJSON_Returns400(t *testing.T) {
	app := newTestApp(&mockConverter{}, &mockExecutor{})

	req := httptest.NewRequest("POST", "/query", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	if resp.StatusCode != 400 {
		t.Errorf("expected 400 for invalid JSON, got %d", resp.StatusCode)
	}
}
