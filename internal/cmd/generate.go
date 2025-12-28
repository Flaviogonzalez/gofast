package cmd

import (
	"path/filepath"

	"github.com/flaviogonzalez/gofast/internal/bootstrap"
	"github.com/flaviogonzalez/gofast/internal/fs"
)

func generate(wd string) error {
	testdir := filepath.Join(wd)
	fs.SetWorkingDir(testdir)

	return bootstrap.GenerateProject(testdir)
}
