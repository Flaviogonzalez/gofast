package bootstrap

import (
	"errors"
	"log"

	"github.com/flaviogonzalez/gofast/internal/fs"
	generator "github.com/flaviogonzalez/gofast/internal/generator"
)

type QueriesGenerationStep struct{}

func (s *QueriesGenerationStep) Name() string { return "Generate SQL queries" }

func (s *QueriesGenerationStep) Run(pctx *ProjectContext) error {
	if pctx.ParsedSchema == nil {
		return errors.New("parsed schema not available")
	}

	for i := range pctx.Tables {
		table := pctx.Tables[i]

		log.Printf("  Table %s has %d columns:\n", table.Name, len(table.Columns))
		for _, col := range table.Columns {
			log.Printf("    - %s (type: %s, auto_increment: %v)\n", col.Name, col.Type, col.AutoIncrement)
		}
		queries, err := generator.GenerateQueriesForTable(*table)
		if err != nil {
			return err
		}

		filename := table.Name + ".sql"
		if err := fs.WriteFileAtQueriesFolder(filename, []byte(queries)); err != nil {
			return err
		}

		log.Printf("  ✓ Generated queries for table: %s\n", table.Name)
	}

	return nil
}
