package bootstrap

import (
	"context"
	"log"

	"github.com/flaviogonzalez/gofast/internal/fs"
	"github.com/flaviogonzalez/gofast/internal/schema"
)

type ProjectContext struct {
	WorkingDir     string
	DatabaseURL    string
	SchemaFilepath string
	ParsedSchema   *schema.Schema  // se llena en el paso de parsing
	Tables         []*schema.Table // alias para ParsedSchema.Tables
	Database       *fs.Database    // se llena en el paso de análisis de la URL
	Ctx            context.Context
}

type Step interface {
	Name() string
	Run(ctx *ProjectContext) error
}

func GenerateProject(wd string) error {
	ctx := context.Background()
	pctx := ProjectContext{
		WorkingDir:     wd,
		DatabaseURL:    "",
		SchemaFilepath: "",
		Ctx:            ctx,
	}
	steps := []Step{
		&ProjectStructureStep{},
		&SchemaInspectionStep{},
		&QueriesGenerationStep{},
		&SQLCConfigStep{},
		&SQLCGenerateStep{},
		&HandlersGenerationStep{},
		&MainFileGenerationStep{},
		&GoModStep{},
	}

	for _, step := range steps {
		log.Printf("→ %s\n", step.Name())
		if err := step.Run(&pctx); err != nil {
			return err
		}
	}

	log.Println("\n🎉 REST API generated successfully!")
	log.Println("\nNext steps:")
	log.Println("  1. go mod tidy")
	log.Println("  2. go run main.go")

	return nil
}
