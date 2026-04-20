package bootstrap

import (
	"errors"

	"github.com/flaviogonzalez/gofast/internal/fs"
	generator "github.com/flaviogonzalez/gofast/internal/generator"
	"github.com/flaviogonzalez/gofast/internal/schema"
)

type RoutesGenerationStep struct{}

func (s *RoutesGenerationStep) Name() string { return "Generate routes.go file" }

func (s *RoutesGenerationStep) Run(pctx *ProjectContext) error {
	if pctx.ParsedSchema == nil {
		return errors.New("parsed schema not available")
	}
	if pctx.Database == nil {
		return errors.New("database info not available")
	}

	tables := make([]schema.Table, len(pctx.Tables))
	for i, t := range pctx.Tables {
		tables[i] = *t
	}

	moduleName := pctx.Database.Database

	routesFile, err := generator.GenerateRoutesFile(tables, moduleName)
	if err != nil {
		return err
	}

	return fs.WriteFileAtHandlersFolder("routes.go", []byte(routesFile))
}
