package text2sql

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fadhlidev/text2sql/cache"
	"github.com/fadhlidev/text2sql/schema"
)

// systemTpl is the instruction template sent to the LLM before every question.
// %s[0] = dialect (e.g. "postgres")
// %s[1] = schema context (tables and columns as plain text)
const systemTpl = `You are a SQL generator for a %s database.
Your goal is to translate natural language questions into valid, read-only SQL queries.

Schema:
%s

Rules:
- SELECT queries only. Never generate INSERT, UPDATE, DELETE, DROP, ALTER, TRUNCATE, or any DDL.
- Use ONLY tables and columns defined in the schema above.
- Always alias tables in JOINs for clarity.
- For unbounded SELECTs, always add a LIMIT 100 clause.
- If the question cannot be answered using the available schema, output exactly: ERROR: <reason>

Output Format:
Respond with a JSON object containing:
"sql": The generated SQL string.
"explanation": A brief, one-sentence natural language explanation of how the query works.

Example:
{
  "sql": "SELECT name FROM users LIMIT 100",
  "explanation": "I am selecting the names of all users, limited to the first 100 results."
}
`

// Translation represents the structured response from the LLM
type Translation struct {
	SQL         string `json:"sql"`
	Explanation string `json:"explanation"`
}

// Converter handles the natural language to SQL translation
type Converter struct {
	llm    LLMClient
	schema *schema.Schema
	cache  cache.Cache
}

// New creates a new Converter instance
func New(llm LLMClient, s *schema.Schema) *Converter {
	return &Converter{
		llm:    llm,
		schema: s,
	}
}

// WithCache attaches a cache to the converter
func (c *Converter) WithCache(ca cache.Cache) *Converter {
	c.cache = ca
	return c
}

// TextToSQL converts a natural language question into a validated SQL query and an explanation
func (c *Converter) TextToSQL(ctx context.Context, question string) (string, string, error) {
	schemaCtx := c.schema.Context()
	
	// Check cache if enabled
	if c.cache != nil {
		key := fmt.Sprintf("sql:%x", sha256.Sum256([]byte(schemaCtx+question)))
		if val, found := c.cache.Get(ctx, key); found {
			// Cache stores "sql|explanation"
			parts := strings.SplitN(val, "|", 2)
			if len(parts) == 2 {
				return parts[0], parts[1], nil
			}
			return val, "", nil
		}
	}

	system := fmt.Sprintf(systemTpl, c.schema.Dialect, schemaCtx)

	resp, err := c.llm.Complete(ctx, system, question)
	if err != nil {
		return "", "", fmt.Errorf("llm complete: %w", err)
	}

	// The LLM might return "ERROR: <reason>"
	if strings.HasPrefix(resp, "ERROR:") {
		return "", "", fmt.Errorf("unanswerable: %s", strings.TrimPrefix(resp, "ERROR:"))
	}

	// Parse JSON response
	var trans Translation
	if err := json.Unmarshal([]byte(resp), &trans); err != nil {
		return "", "", fmt.Errorf("llm returned invalid json: %w", err)
	}

	sql := strings.TrimSpace(trans.SQL)
	if err := validate(sql); err != nil {
		return "", "", fmt.Errorf("validation failed: %w", err)
	}

	// Store in cache if enabled
	if c.cache != nil {
		key := fmt.Sprintf("sql:%x", sha256.Sum256([]byte(schemaCtx+question)))
		c.cache.Set(ctx, key, sql+"|"+trans.Explanation, 24*time.Hour)
	}

	return sql, trans.Explanation, nil
}
