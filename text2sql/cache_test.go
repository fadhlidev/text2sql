package text2sql

import (
	"context"
	"testing"

	"github.com/fadhlidev/text2sql/cache"
	"github.com/fadhlidev/text2sql/schema"
)

type cacheMockLLM struct {
	count int
}

func (m *cacheMockLLM) Complete(ctx context.Context, system, user string) (string, error) {
	m.count++
	return `{"sql": "SELECT * FROM users", "explanation": "test"}`, nil
}

func TestConverter_Cache(t *testing.T) {
	llm := &cacheMockLLM{}
	s := &schema.Schema{Dialect: "postgres"}
	ca := cache.NewMemoryCache(0, 0)
	conv := New(llm, s).WithCache(ca)

	ctx := context.Background()
	q := "get all users"

	// 1. First call - should call LLM
	_, _, err := conv.TextToSQL(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	if llm.count != 1 {
		t.Errorf("expected 1 LLM call, got %d", llm.count)
	}

	// 2. Second call - should be cached
	_, expl, err := conv.TextToSQL(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	if llm.count != 1 {
		t.Errorf("expected still 1 LLM call (cached), got %d", llm.count)
	}
	if expl != "test" {
		t.Errorf("expected explanation 'test' from cache, got: %s", expl)
	}

	// 3. Different question - should call LLM
	_, _, err = conv.TextToSQL(ctx, "get admins")
	if err != nil {
		t.Fatal(err)
	}
	if llm.count != 2 {
		t.Errorf("expected 2 LLM calls, got %d", llm.count)
	}
}
