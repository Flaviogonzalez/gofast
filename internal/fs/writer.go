package fs

import (
	"os"
	"path/filepath"
)

var WorkingDir string
var SQLFolder = filepath.Join(WorkingDir, "sql")
var SchemaFolder = filepath.Join(SQLFolder, "schema")
var QueriesFolder = filepath.Join(SQLFolder, "queries")
var InternalFolder = filepath.Join(WorkingDir, "internal")
var HandlersFolder = filepath.Join(InternalFolder, "handlers")
var ModelsFolder = filepath.Join(InternalFolder, "models")
var CmdFolder = filepath.Join(WorkingDir, "cmd")
var ApiFolder = filepath.Join(CmdFolder, "api")

func CreateProjectStructure() error {
	dirs := []string{
		SQLFolder,
		SchemaFolder,
		QueriesFolder,
		InternalFolder,
		HandlersFolder,
		ModelsFolder,
		CmdFolder,
		ApiFolder,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

func WriteFile(path string, content []byte) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}

func WriteFileAtSQLFolder(filename string, content []byte) error {
	path := filepath.Join(SQLFolder, filename)
	return WriteFile(path, content)
}

func WriteFileAtSchemaFolder(filename string, content []byte) error {
	path := filepath.Join(SchemaFolder, filename)
	return WriteFile(path, content)
}

func WriteFileAtQueriesFolder(filename string, content []byte) error {
	path := filepath.Join(QueriesFolder, filename)
	return WriteFile(path, content)
}

func WriteFileAtHandlersFolder(filename string, content []byte) error {
	path := filepath.Join(HandlersFolder, filename)
	return WriteFile(path, content)
}

func WriteFileAtModelsFolder(filename string, content []byte) error {
	path := filepath.Join(ModelsFolder, filename)
	return WriteFile(path, content)
}

func SetWorkingDir(wd string) {
	WorkingDir = wd
	// recalcular todas las rutas si es necesario
	SQLFolder = filepath.Join(WorkingDir, "sql")
	SchemaFolder = filepath.Join(SQLFolder, "schema")
	QueriesFolder = filepath.Join(SQLFolder, "queries")
	InternalFolder = filepath.Join(WorkingDir, "internal")
	HandlersFolder = filepath.Join(InternalFolder, "handlers")
	ModelsFolder = filepath.Join(InternalFolder, "models")
}
