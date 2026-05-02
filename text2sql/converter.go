package text2sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/fadhlidev/text2sql/schema"
)

// systemTpl is the instruction template sent to the LLM before every question.
// %s[0] = dialect (e.g. "postgres")
// %s[1] = schema context (tables and columns as plain text)
const systemTpl = `You are a SQL generator for a %s database.
Output ONLY valid SQL — no explanation, no markdown, no backticks.

Rules:
- SELECT queries only. Never generate INSERT, UPDATE, DELETE, DROP, ALTER, TRUNCATE, or any DDL.
- Use ONLY tables and columns defined in the schema below.
- Always alias tables in JOINs for clarity.
- For unbounded SELECTs, always add a LIMIT 100 clause.
- If the question cannot be answered using the available schema, output exactly: ERROR: <reason>

Schema:
%s`

// Converter handles the natural language to SQL translation
type Converter struct {
	llm    LLMClient
	schema *schema.Schema
}

// New creates a new Converter instance
func New(llm LLMClient, s *schema.Schema) *Converter {
	return &Converter{
		llm:    llm,
		schema: s,
	}
}

// TextToSQL converts a natural language question into a validated SQL query
func (c *Converter) TextToSQL(ctx context.Context, question string) (string, error) {
	system := fmt.Sprintf(systemTpl, c.schema.Dialect, c.schema.Context())

	resp, err := c.llm.Complete(ctx, system, question)
	if err != nil {
		return "", fmt.Errorf("llm complete: %w", err)
	}

	// The LLM might return "ERROR: <reason>"
	if strings.HasPrefix(resp, "ERROR:") {
		return "", fmt.Errorf("unanswerable: %s", strings.TrimPrefix(resp, "ERROR:"))
	}

	// Clean up markdown if the LLM ignored the "no markdown" rule
	sql := strings.TrimSpace(resp)
	sql = strings.TrimPrefix(sql, "```sql")
	sql = strings.TrimPrefix(sql, "```")
	sql = strings.TrimSuffix(sql, "```")
	sql = strings.TrimSpace(sql)

	if err := validate(sql); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	return sql, nil
}
