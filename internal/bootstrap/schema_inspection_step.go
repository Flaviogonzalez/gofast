// internal/bootstrap/schema_step.go
package bootstrap

import (
	"bytes"
	"log"
	"os/exec"
	"path/filepath"

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

	// Analyze database URL to get database info
	dbInfo, err := fs.AnalyzeDatabaseURL(dbURL)
	if err != nil {
		return err
	}
	pctx.Database = dbInfo

	var out bytes.Buffer
	cmd := exec.CommandContext(pctx.Ctx, "atlas", "schema", "inspect", "-u", pctx.DatabaseURL, "--format", "{{ sql . }}")
	cmd.Stdout = &out
	cmd.Stderr = log.Writer()

	if err := cmd.Run(); err != nil {
		return err
	}

	if err := fs.WriteFileAtSchemaFolder("schema.sql", out.Bytes()); err != nil {
		return err
	}
	pctx.SchemaFilepath = filepath.Join(fs.SchemaFolder, "schema.sql")

	loader := schema.NewSQLLoader(pctx.SchemaFilepath)
	parsed, err := loader.Load()
	if err != nil {
		return err
	}

	pctx.ParsedSchema = parsed
	tables := make([]*schema.Table, len(parsed.Tables))
	for i := range parsed.Tables {
		tables[i] = &parsed.Tables[i]
	}
	pctx.Tables = tables

	log.Println("✓ Schema saved to sql/schema/schema.sql")
	return nil
}
