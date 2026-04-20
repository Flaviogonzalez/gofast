package bootstrap

import (
	"errors"
	"log"

	"github.com/flaviogonzalez/gofast/internal/fs"
	"github.com/flaviogonzalez/gofast/internal/generator"
)

type ModelsGenerationStep struct{}

func (s *ModelsGenerationStep) Name() string { return "Generate Go models" }

func (s *ModelsGenerationStep) Run(pctx *ProjectContext) error {
	if pctx.ParsedSchema == nil {
		return errors.New("parsed schema not available")
	}
	if len(pctx.Tables) == 0 {
		log.Println("  ⚠ no tables to generate models for — skipping")
		return nil
	}

	// Generate db.go (DBTX interface + Queries struct)
	dbFile, err := generator.GenerateDBFile()
	if err != nil {
		return err
	}
	if err := fs.WriteFileAtModelsFolder("db.go", []byte(dbFile)); err != nil {
		return err
	}
	log.Println("  ✓ Generated db.go")

	// Generate models.go (all structs)
	tables := tablesToValues(pctx.Tables)
	modelsFile, err := generator.GenerateModelsFile(tables)
	if err != nil {
		return err
	}
	if err := fs.WriteFileAtModelsFolder("models.go", []byte(modelsFile)); err != nil {
		return err
	}
	log.Println("  ✓ Generated models.go")

	// Generate query functions per table
	for _, table := range pctx.Tables {
		queryFile, err := generator.GenerateQueryFuncsForTable(*table)
		if err != nil {
			return err
		}
		filename := table.Name + ".sql.go"
		if err := fs.WriteFileAtModelsFolder(filename, []byte(queryFile)); err != nil {
			return err
		}
		log.Printf("  ✓ Generated %s\n", filename)
	}

	return nil
}
