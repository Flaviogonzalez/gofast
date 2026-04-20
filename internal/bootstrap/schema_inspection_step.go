// internal/bootstrap/schema_step.go
package bootstrap

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

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

	if _, lookErr := exec.LookPath("atlas"); lookErr != nil {
		return fmt.Errorf("atlas CLI not found in PATH — install it from https://atlasgo.io/getting-started")
	}

	var out, errBuf bytes.Buffer
	cmd := exec.CommandContext(pctx.Ctx, "atlas", "schema", "inspect", "-u", pctx.DatabaseURL, "--format", "{{ sql . }}")
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("atlas schema inspect failed: %w\n%s", err, strings.TrimSpace(errBuf.String()))
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

func hasIDColumn(t *schema.Table) bool {
	for _, c := range t.Columns {
		if strings.EqualFold(c.Name, "id") {
			return true
		}
	}
	return false
}
