package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/flaviogonzalez/gofast/internal/fs"
	"github.com/flaviogonzalez/gofast/internal/schema"
)

type SchemaInspectionStep struct{}

func (s *SchemaInspectionStep) Name() string { return "Inspect database schema" }

func (s *SchemaInspectionStep) Run(pctx *ProjectContext) error {
	dbURL, err := fs.ReadDatabaseURL()
	if err != nil {
		return err
	}
	pctx.DatabaseURL = dbURL

	dbInfo, err := fs.AnalyzeDatabaseURL(dbURL)
	if err != nil {
		return err
	}
	pctx.Database = dbInfo

	dsn, err := fs.ToMySQLDSN(dbURL)
	if err != nil {
		return fmt.Errorf("cannot convert DATABASE_URL to DSN: %w", err)
	}

	ddl, err := inspectMySQL(pctx.Ctx, dsn)
	if err != nil {
		return fmt.Errorf("schema inspection failed: %w", err)
	}

	if err := fs.WriteFileAtSchemaFolder("schema.sql", []byte(ddl)); err != nil {
		return err
	}
	pctx.SchemaFilepath = filepath.Join(fs.SchemaFolder, "schema.sql")

	loader := schema.NewSQLLoader(pctx.SchemaFilepath)
	parsed, err := loader.Load()
	if err != nil {
		return err
	}

	pctx.ParsedSchema = parsed
	tables := make([]*schema.Table, 0, len(parsed.Tables))
	for i := range parsed.Tables {
		t := &parsed.Tables[i]
		if !hasIDColumn(t) {
			log.Printf("  ⚠ skipping table %q: no `id` column (tables without an id are not supported yet)", t.Name)
			continue
		}
		tables = append(tables, t)
	}
	pctx.Tables = tables

	log.Println("✓ Schema saved to sql/schema/schema.sql")
	return nil
}

// inspectMySQL connects to the database and retrieves CREATE TABLE statements
// for every table, producing the same DDL that `atlas schema inspect --format "{{ sql . }}"` would.
func inspectMySQL(ctx context.Context, dsn string) (string, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return "", err
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return "", fmt.Errorf("cannot reach database: %w", err)
	}

	rows, err := db.QueryContext(ctx, "SHOW TABLES")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var tableNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return "", err
		}
		tableNames = append(tableNames, name)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	var ddl strings.Builder
	for _, name := range tableNames {
		row := db.QueryRowContext(ctx, "SHOW CREATE TABLE `"+name+"`")
		var tblName, createStmt string
		if err := row.Scan(&tblName, &createStmt); err != nil {
			return "", fmt.Errorf("SHOW CREATE TABLE %s: %w", name, err)
		}
		ddl.WriteString(createStmt)
		ddl.WriteString(";\n")
	}
	return ddl.String(), nil
}

func hasIDColumn(t *schema.Table) bool {
	for _, c := range t.Columns {
		if strings.EqualFold(c.Name, "id") {
			return true
		}
	}
	return false
}

// tablesToValues converts a slice of table pointers to table values.
func tablesToValues(tables []*schema.Table) []schema.Table {
	result := make([]schema.Table, len(tables))
	for i, t := range tables {
		result[i] = *t
	}
	return result
}
