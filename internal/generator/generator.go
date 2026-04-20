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

//go:embed templates/handlers.go.tmpl
var handlersTemplate string

//go:embed templates/main.go.tmpl
var mainTemplate string

//go:embed templates/routes.go.tmpl
var routesTemplate string

//go:embed templates/db.go.tmpl
var dbTemplate string

//go:embed templates/models.go.tmpl
var modelsTemplate string

//go:embed templates/query_funcs.go.tmpl
var queryFuncsTemplate string

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

// ModelsData holds data for models.go generation
type ModelsData struct {
	Tables []schema.Table
}

var funcMap = template.FuncMap{
	"add":              add,
	"singular":         singular,
	"title":            strings.Title,
	"toPascal":         toPascal,
	"updatableColumns": updatableColumns,
	"idGoType":         idGoType,
	"idParseFunc":      idParseFunc,
	"idBitSize":        idBitSize,
	"columnGoType":     columnGoType,
	"needsTimeImport":  needsTimeImport,
	"columnList":       columnList,
	"scanFields":       scanFields,
	"paramFields":      paramFields,
	"placeholders":     placeholders,
	"updateSetClauses": updateSetClauses,
	"updateParamFields": updateParamFields,
	"insColList":        insColList,
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

// toPascal converts a snake_case identifier to PascalCase.
// Special abbreviations (id, url, api, sql, http) are uppercased.
func toPascal(s string) string {
	if s == "" {
		return s
	}
	abbrevs := map[string]string{
		"id": "ID", "url": "URL", "api": "API",
		"sql": "SQL", "http": "HTTP", "ip": "IP",
	}
	parts := strings.Split(s, "_")
	for i, p := range parts {
		lower := strings.ToLower(p)
		if up, ok := abbrevs[lower]; ok {
			parts[i] = up
		} else if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

// columnGoType returns the Go type for a column, using pointers for nullable columns.
func columnGoType(c schema.Column) string {
	base := baseGoType(c)
	if c.Nullable {
		return "*" + base
	}
	return base
}

// baseGoType returns the non-pointer Go type for a column.
func baseGoType(c schema.Column) string {
	switch c.Type {
	case schema.TypeInt:
		return goIntType(c)
	case schema.TypeFloat:
		raw := strings.ToUpper(c.RawType)
		switch {
		case strings.Contains(raw, "DOUBLE"), strings.Contains(raw, "NUMERIC"),
			strings.Contains(raw, "DECIMAL"), strings.Contains(raw, "DEC"):
			return "float64"
		default:
			return "float32"
		}
	case schema.TypeBool:
		return "bool"
	case schema.TypeTime:
		return "time.Time"
	case schema.TypeString:
		return "string"
	default:
		return "string"
	}
}

// needsTimeImport returns true if any column in the tables uses time.Time.
func needsTimeImport(tables []schema.Table) bool {
	for _, t := range tables {
		for _, c := range t.Columns {
			if c.Type == schema.TypeTime {
				return true
			}
		}
	}
	return false
}

// columnList returns a comma-separated, backtick-quoted list of all column names.
func columnList(t schema.Table) string {
	cols := make([]string, len(t.Columns))
	for i, c := range t.Columns {
		cols[i] = "`" + c.Name + "`"
	}
	return strings.Join(cols, ", ")
}

// scanFields returns "&i.Field1, &i.Field2, ..." for row.Scan.
func scanFields(t schema.Table) string {
	fields := make([]string, len(t.Columns))
	for i, c := range t.Columns {
		fields[i] = "&i." + toPascal(c.Name)
	}
	return strings.Join(fields, ", ")
}

// paramFields returns "arg.Field1, arg.Field2, ..." for ExecContext args (non-auto-increment columns).
func paramFields(t schema.Table) string {
	var fields []string
	for _, c := range t.Columns {
		if !c.AutoIncrement {
			fields = append(fields, "arg."+toPascal(c.Name))
		}
	}
	return strings.Join(fields, ", ")
}

// placeholders returns "?, ?, ..." for non-auto-increment columns.
func placeholders(t schema.Table) string {
	var count int
	for _, c := range t.Columns {
		if !c.AutoIncrement {
			count++
		}
	}
	return strings.Join(repeatStr("?", count), ", ")
}

func repeatStr(s string, n int) []string {
	result := make([]string, n)
	for i := range result {
		result[i] = s
	}
	return result
}

// updateSetClauses returns "`col1` = ?, `col2` = ?" for non-auto-increment columns.
func updateSetClauses(t schema.Table) string {
	var clauses []string
	for _, c := range t.Columns {
		if !c.AutoIncrement {
			clauses = append(clauses, "`"+c.Name+"` = ?")
		}
	}
	return strings.Join(clauses, ", ")
}

// updateParamFields returns "arg.Field1, arg.Field2, ..., arg.ID" for UPDATE ExecContext args.
func updateParamFields(t schema.Table) string {
	var fields []string
	for _, c := range t.Columns {
		if !c.AutoIncrement {
			fields = append(fields, "arg."+toPascal(c.Name))
		}
	}
	fields = append(fields, "arg.ID")
	return strings.Join(fields, ", ")
}

// insColList returns a backtick-quoted list of non-auto-increment columns for INSERT.
func insColList(t schema.Table) string {
	var cols []string
	for _, c := range t.Columns {
		if !c.AutoIncrement {
			cols = append(cols, "`"+c.Name+"`")
		}
	}
	return strings.Join(cols, ", ")
}

// GenerateModelsFile generates the models.go file with all structs.
func GenerateModelsFile(tables []schema.Table) (string, error) {
	tmpl, err := template.New("models").Funcs(funcMap).Parse(modelsTemplate)
	if err != nil {
		return "", err
	}

	data := ModelsData{Tables: tables}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateDBFile generates the db.go file with DBTX interface and Queries struct.
func GenerateDBFile() (string, error) {
	tmpl, err := template.New("db").Parse(dbTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GenerateQueryFuncsForTable generates the query functions file for a single table.
func GenerateQueryFuncsForTable(table schema.Table) (string, error) {
	fmap := template.FuncMap{
		"add":               add,
		"singular":          singular,
		"title":             strings.Title,
		"toPascal":          toPascal,
		"updatableColumns":  updatableColumns,
		"idGoType":          idGoType,
		"idParseFunc":       idParseFunc,
		"idBitSize":         idBitSize,
		"columnGoType":      columnGoType,
		"needsTimeImport":   needsTimeImport,
		"columnList":        columnList,
		"scanFields":        scanFields,
		"paramFields":       paramFields,
		"placeholders":      placeholders,
		"updateSetClauses":  updateSetClauses,
		"updateParamFields": updateParamFields,
		"insColList":        insColList,
	}
	tmpl, err := template.New("query_funcs").Funcs(fmap).Parse(queryFuncsTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, table); err != nil {
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
