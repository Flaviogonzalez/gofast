package cmd

import (
	"path/filepath"

	"github.com/flaviogonzalez/gofast/internal/bootstrap"
	"github.com/flaviogonzalez/gofast/internal/fs"
)

func generate(wd string) error {
	// Opcional: establecer wd global si fs lo necesita
	testdir := filepath.Join(wd, "dist")
	fs.SetWorkingDir(testdir)

	return bootstrap.GenerateProject(testdir)
}
