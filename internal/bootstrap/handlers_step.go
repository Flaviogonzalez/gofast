package bootstrap

import (
	"errors"
	"log"

	"github.com/flaviogonzalez/gofast/internal/fs"
	generator "github.com/flaviogonzalez/gofast/internal/generator"
)

type HandlersGenerationStep struct {
	SchemaPath string
}

func (s *HandlersGenerationStep) Name() string { return "Generate HTTP handlers" }

func (s *HandlersGenerationStep) Run(pctx *ProjectContext) error {
	if pctx.ParsedSchema == nil {
		return errors.New("parsed schema not available")
	}
	if pctx.Database == nil {
		return errors.New("database info not available")
	}

	// Use database name as module name
	moduleName := pctx.Database.Database

	for i := range pctx.Tables {
		table := pctx.Tables[i]

		handlers, err := generator.GenerateHandlersForTable(*table, moduleName)
		if err != nil {
			return err
		}

		filename := table.Name + "_handler.go"
		if err := fs.WriteFileAtHandlersFolder(filename, []byte(handlers)); err != nil {
			return err
		}

		log.Printf("  ✓ Generated handlers for table: %s\n", table.Name)
	}

	return nil
}
