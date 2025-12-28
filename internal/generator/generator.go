package generator

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"

	"github.com/flaviogonzalez/gofast/internal/schema"
)

//go:embed templates/queries.sql.tmpl
var queriesTemplate string

//go:embed templates/sqlc.yaml.tmpl
var sqlcConfigTemplate string

//go:embed templates/handlers.go.tmpl
var handlersTemplate string

//go:embed templates/main.go.tmpl
var mainTemplate string

// HandlerData holds data for handler generation
type HandlerData struct {
	Table      schema.Table
	ModuleName string
}

// MainFileData holds data for main.go generation
type MainFileData struct {
	Tables     []schema.Table
	ModuleName string
	Driver     string
}

var funcMap = template.FuncMap{
	"add":              add,
	"singular":         singular,
	"title":            strings.Title,
	"updatableColumns": updatableColumns,
}

func GenerateQueriesForTable(table schema.Table) (string, error) {

	tmpl, err := template.New("queries").Funcs(funcMap).Parse(queriesTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, table); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func add(a, b int) int {
	return a + b
}

func singular(s string) string {
	if strings.HasSuffix(s, "s") {
		return s[:len(s)-1]
	}
	return s
}

func updatableColumns(t schema.Table) []schema.Column {
	var cols []schema.Column
	for _, c := range t.Columns {
		if !c.AutoIncrement {
			cols = append(cols, c)
		}
	}
	return cols
}

func CreateSQLCConfig() (string, error) {
	tmpl, err := template.New("sqlc").Parse(sqlcConfigTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func GenerateHandlersForTable(table schema.Table, moduleName string) (string, error) {
	tmpl, err := template.New("handlers").Funcs(funcMap).Parse(handlersTemplate)
	if err != nil {
		return "", err
	}

	data := HandlerData{
		Table:      table,
		ModuleName: moduleName,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func GenerateMainFile(tables []schema.Table, moduleName, driver string) (string, error) {
	tmpl, err := template.New("main").Funcs(funcMap).Parse(mainTemplate)
	if err != nil {
		return "", err
	}

	data := MainFileData{
		Tables:     tables,
		ModuleName: moduleName,
		Driver:     driver,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
