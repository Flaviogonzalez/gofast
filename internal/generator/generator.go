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

//go:embed templates/routes.go.tmpl
var routesTemplate string

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
	"idGoType":         idGoType,
	"idParseFunc":      idParseFunc,
	"idBitSize":        idBitSize,
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

// singular converts a (possibly plural) identifier to its singular form using a
// small set of English rules. It is intentionally simple; callers with irregular
// plurals should feed already-singular table names.
func singular(s string) string {
	if s == "" {
		return s
	}
	lower := strings.ToLower(s)

	// -ies → -y (categories → category). Guard short tokens like "lies".
	if strings.HasSuffix(lower, "ies") && len(s) > 3 {
		return s[:len(s)-3] + "y"
	}
	// Double-letter endings that look plural but aren't.
	for _, keep := range []string{"ss", "us", "is", "os"} {
		if strings.HasSuffix(lower, keep) {
			return s
		}
	}
	// -es endings that need the "es" dropped (addresses → address, boxes → box).
	for _, suf := range []string{"sses", "xes", "zes", "ches", "shes"} {
		if strings.HasSuffix(lower, suf) {
			return s[:len(s)-2]
		}
	}
	// Plain trailing -s.
	if strings.HasSuffix(lower, "s") {
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

// findIDColumn returns the `id` column for a table if present, otherwise the
// first AUTO_INCREMENT column, otherwise a zero-value column.
func findIDColumn(t schema.Table) schema.Column {
	for _, c := range t.Columns {
		if strings.EqualFold(c.Name, "id") {
			return c
		}
	}
	for _, c := range t.Columns {
		if c.AutoIncrement {
			return c
		}
	}
	return schema.Column{}
}

// idGoType returns the Go type (matching what sqlc will emit) for the table's
// primary id column. Falls back to int64 when the column can't be located.
func idGoType(t schema.Table) string {
	c := findIDColumn(t)
	return goIntType(c)
}

// idParseFunc returns the strconv function name used to parse the URL {id}
// into the matching Go integer type.
func idParseFunc(t schema.Table) string {
	c := findIDColumn(t)
	if c.Unsigned {
		return "ParseUint"
	}
	return "ParseInt"
}

// idBitSize returns the bitSize argument passed to strconv.Parse{Int,Uint}.
func idBitSize(t schema.Table) int {
	c := findIDColumn(t)
	switch strings.ToUpper(c.RawType) {
	case "TINYINT":
		return 8
	case "SMALLINT":
		return 16
	case "MEDIUMINT", "INT", "INTEGER", "INT4":
		return 32
	case "BIGINT", "INT8", "SERIAL", "BIGSERIAL":
		return 64
	}
	return 64
}

func goIntType(c schema.Column) string {
	prefix := "int"
	if c.Unsigned {
		prefix = "uint"
	}
	switch strings.ToUpper(c.RawType) {
	case "TINYINT":
		return prefix + "8"
	case "SMALLINT", "INT2":
		return prefix + "16"
	case "MEDIUMINT", "INT", "INTEGER", "INT4":
		return prefix + "32"
	case "BIGINT", "INT8":
		return prefix + "64"
	case "SERIAL", "BIGSERIAL":
		return "int64"
	}
	return "int64"
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

func GenerateRoutesFile(tables []schema.Table, moduleName string) (string, error) {
	tmpl, err := template.New("routes").Funcs(funcMap).Parse(routesTemplate)
	if err != nil {
		return "", err
	}

	data := MainFileData{
		Tables:     tables,
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
