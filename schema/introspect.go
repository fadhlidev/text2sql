package schema

import (
	"context"
	"database/sql"
	"fmt"
)

// Introspect reads the schema from the connected database.
// dialect should be "postgres", "mysql", or "sqlite".
func Introspect(ctx context.Context, db *sql.DB, dialect string) (*Schema, error) {
	switch dialect {
	case "postgres":
		return introspectPostgres(ctx, db)
	case "mysql":
		return introspectMySQL(ctx, db)
	case "sqlite":
		return introspectSQLite(ctx, db)
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}
}
