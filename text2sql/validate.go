package text2sql

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// onlySelect matches queries that start with SELECT or WITH (for CTEs like "WITH x AS (...) SELECT ...")
var onlySelect = regexp.MustCompile(`(?i)^\s*(WITH|SELECT)\s`)

// blocked is a list of keywords that must never appear in generated SQL
var blocked = []string{
	"INSERT", "UPDATE", "DELETE", "DROP", "ALTER",
	"TRUNCATE", "CREATE", "GRANT", "REVOKE",
	"EXEC", "EXECUTE", "--", "/*",
}

// validate checks that a SQL string is safe to execute.
// Returns an error if the query is not a SELECT or contains dangerous keywords.
func validate(sql string) error {
	if !onlySelect.MatchString(sql) {
		return errors.New("only SELECT/WITH queries are permitted")
	}

	upper := strings.ToUpper(sql)
	for _, kw := range blocked {
		if strings.Contains(upper, kw) {
			return fmt.Errorf("forbidden keyword detected: %s", kw)
		}
	}

	return nil
}
