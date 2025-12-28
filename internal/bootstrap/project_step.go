package bootstrap

import (
	"github.com/flaviogonzalez/gofast/internal/fs"
)

type ProjectStructureStep struct{}

func (s *ProjectStructureStep) Name() string { return "Generate project structure and files" }
func (s *ProjectStructureStep) Run(ctx *ProjectContext) error {
	return fs.CreateProjectStructure()
}
