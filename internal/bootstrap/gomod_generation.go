package bootstrap

import (
	"fmt"
	"path/filepath"

	"github.com/flaviogonzalez/gofast/internal/fs"
)

type GoModStep struct {
	WorkingDir string
}

func (s *GoModStep) Name() string { return "Generate go.mod file" }

func (s *GoModStep) Run(pctx *ProjectContext) error {
	moduleName := fmt.Sprintf("github.com/%s", pctx.Database.Database)

	goModContent := fmt.Sprintf(`module %s

go 1.23

require (
	github.com/go-chi/chi/v5 v5.2.0
	github.com/go-sql-driver/mysql v1.8.1
)

require filippo.io/edwards25519 v1.1.0 // indirect
`, moduleName)

	goModPath := filepath.Join(pctx.WorkingDir, "go.mod")
	return fs.WriteFile(goModPath, []byte(goModContent))
}
