package bootstrap

import (
	"log"
	"path/filepath"

	"github.com/flaviogonzalez/gofast/internal/fs"
	"github.com/flaviogonzalez/gofast/internal/generator"
)

type SQLCConfigStep struct{}

func (s *SQLCConfigStep) Name() string { return "Generate sqlc configuration" }

func (s *SQLCConfigStep) Run(pctx *ProjectContext) error {
	if pctx.ParsedSchema == nil || len(pctx.Tables) == 0 {
		return nil
	}

	configContent, err := generator.CreateSQLCConfig()
	if err != nil {
		return err
	}

	sqlcPath := filepath.Join(fs.WorkingDir, "sqlc.yaml")
	if err := fs.WriteFile(sqlcPath, []byte(configContent)); err != nil {
		return err
	}

	log.Println("  ✓ sqlc.yaml generated")
	return nil
}
