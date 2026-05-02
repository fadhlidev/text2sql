package schema

import (
	"fmt"
	"strings"
)

// Column represents one column in a database table
type Column struct {
	Name     string // e.g. "customer_id"
	Type     string // e.g. "integer", "varchar", "timestamp"
	Nullable bool   // true if the column can be NULL
	IsPK     bool   // true if this is the primary key
	FKRef    string // if this is a foreign key, e.g. "customers.id"
}

// Table represents one database table with its columns
type Table struct {
	Name    string
	Columns []Column
}

// Schema represents the entire database structure
type Schema struct {
	Dialect string // "postgres", "mysql", "sqlite"
	Tables  []Table
}

// Context converts the schema to a compact text string for injection into LLM prompts.
// Example output:
//   Table: orders
//     id integer [PK] NOT NULL
//     customer_id integer [FK→customers.id] NOT NULL
//     total numeric
func (s *Schema) Context() string {
	var sb strings.Builder
	for _, t := range s.Tables {
		fmt.Fprintf(&sb, "Table: %s\n", t.Name)
		for _, c := range t.Columns {
			meta := ""
			if c.IsPK {
				meta += " [PK]"
			}
			if c.FKRef != "" {
				meta += " [FK→" + c.FKRef + "]"
			}
			if !c.Nullable {
				meta += " NOT NULL"
			}
			fmt.Fprintf(&sb, "  %s %s%s\n", c.Name, c.Type, meta)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
