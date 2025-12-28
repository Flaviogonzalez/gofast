package bootstrap

import (
	"errors"
	"path/filepath"

	"github.com/flaviogonzalez/gofast/internal/fs"
	generator "github.com/flaviogonzalez/gofast/internal/generator"
	"github.com/flaviogonzalez/gofast/internal/schema"
)

type MainFileGenerationStep struct{}

func (s *MainFileGenerationStep) Name() string { return "Generate main.go file" }

func (s *MainFileGenerationStep) Run(pctx *ProjectContext) error {
	if pctx.ParsedSchema == nil {
		return errors.New("parsed schema not available")
	}
	if pctx.Database == nil {
		return errors.New("database info not available")
	}

	// Convert table pointers to values
	tables := make([]schema.Table, len(pctx.Tables))
	for i, t := range pctx.Tables {
		tables[i] = *t
	}

	moduleName := pctx.Database.Database
	driver := pctx.Database.Driver

	mainFile, err := generator.GenerateMainFile(tables, moduleName, driver)
	if err != nil {
		return err
	}

	mainPath := filepath.Join(pctx.WorkingDir, "cmd", "api", "main.go")
	if err := fs.WriteFile(mainPath, []byte(mainFile)); err != nil {
		return err
	}

	return nil
}
