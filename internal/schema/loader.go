package schema

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

type SQLLoader struct {
	filepath string
}

func NewSQLLoader(filepath string) *SQLLoader {
	return &SQLLoader{filepath: filepath}
}

func (l *SQLLoader) Load() (*Schema, error) {
	file, err := os.Open(l.filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	schema := &Schema{Tables: []Table{}}
	scanner := bufio.NewScanner(file)

	var currentTable *Table
	inCreateTable := false

	createTableRe := regexp.MustCompile(`(?i)CREATE\s+(?:TEMPORARY\s+)?TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?[\x60"']?(\w+)[\x60"']?\s*\(`)
	columnRe := regexp.MustCompile(`^\s*[\x60"'](\w+)[\x60"']\s+(\w+(?:\([^)]*\))?(?:\s+(?:UNSIGNED|SIGNED|ZEROFILL))?)(.*)`)

	for scanner.Scan() {
		line := scanner.Text()
		
		// Remove inline comments
		if idx := strings.Index(line, "--"); idx >= 0 {
			line = line[:idx]
		}
		
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Detect CREATE TABLE
		if matches := createTableRe.FindStringSubmatch(trimmed); matches != nil {
			inCreateTable = true
			currentTable = &Table{Name: matches[1], Columns: []Column{}}
			continue
		}

		// Detect end of CREATE TABLE
		if inCreateTable && (strings.HasPrefix(trimmed, ")") || strings.Contains(trimmed, ") ENGINE") || 
		    strings.Contains(trimmed, ") CHARSET") || strings.Contains(trimmed, ") COLLATE") ||
			strings.Contains(trimmed, ") AUTO_INCREMENT")) {
			if currentTable != nil {
				schema.Tables = append(schema.Tables, *currentTable)
			}
			inCreateTable = false
			currentTable = nil
			continue
		}

		// Process columns
		if inCreateTable && currentTable != nil {
			// Skip constraint definitions
			upperTrimmed := strings.ToUpper(trimmed)
			if strings.HasPrefix(upperTrimmed, "PRIMARY") ||
				strings.HasPrefix(upperTrimmed, "FOREIGN") ||
				strings.HasPrefix(upperTrimmed, "UNIQUE") ||
				strings.HasPrefix(upperTrimmed, "INDEX") ||
				strings.HasPrefix(upperTrimmed, "KEY") ||
				strings.HasPrefix(upperTrimmed, "CONSTRAINT") ||
				strings.HasPrefix(upperTrimmed, "CHECK") ||
				strings.HasPrefix(upperTrimmed, "FULLTEXT") ||
				strings.HasPrefix(upperTrimmed, "SPATIAL") {
				continue
			}

			if matches := columnRe.FindStringSubmatch(line); matches != nil {
				colName := matches[1]
				colTypeRaw := strings.TrimSpace(matches[2])
				remainder := strings.ToUpper(strings.TrimSpace(matches[3]))

				// Extract base type
				colTypeBase := strings.ToUpper(colTypeRaw)
				if idx := strings.Index(colTypeBase, "("); idx != -1 {
					colTypeBase = strings.TrimSpace(colTypeBase[:idx])
				}

				// Parse column attributes
				nullable := !strings.Contains(remainder, "NOT NULL")
				autoIncrement := strings.Contains(remainder, "AUTO_INCREMENT") ||
					strings.Contains(remainder, "AUTOINCREMENT")

				column := Column{
					Name:          colName,
					Type:          mapSQLTypeToGoType(colTypeBase),
					Nullable:      nullable,
					AutoIncrement: autoIncrement,
				}

				currentTable.Columns = append(currentTable.Columns, column)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return schema, nil
}

func mapSQLTypeToGoType(sqlType string) ColumnType {
	sqlType = strings.ToUpper(strings.TrimSpace(sqlType))

	switch {
	// Tipos enteros
	case strings.Contains(sqlType, "BIGINT"), strings.Contains(sqlType, "INT8"):
		return TypeInt
	case strings.Contains(sqlType, "INTEGER"), strings.Contains(sqlType, "INT"),
		strings.Contains(sqlType, "INT4"), strings.Contains(sqlType, "SMALLINT"),
		strings.Contains(sqlType, "TINYINT"), strings.Contains(sqlType, "MEDIUMINT"),
		strings.Contains(sqlType, "INT2"):
		return TypeInt
	case strings.Contains(sqlType, "SERIAL"), strings.Contains(sqlType, "BIGSERIAL"):
		return TypeInt

	// Tipos de cadena
	case strings.Contains(sqlType, "VARCHAR"), strings.Contains(sqlType, "CHAR"),
		strings.Contains(sqlType, "NVARCHAR"), strings.Contains(sqlType, "NCHAR"):
		return TypeString
	case strings.Contains(sqlType, "TEXT"), strings.Contains(sqlType, "TINYTEXT"),
		strings.Contains(sqlType, "MEDIUMTEXT"), strings.Contains(sqlType, "LONGTEXT"),
		strings.Contains(sqlType, "CLOB"):
		return TypeString
	case strings.Contains(sqlType, "ENUM"), strings.Contains(sqlType, "SET"):
		return TypeString
	case strings.Contains(sqlType, "JSON"), strings.Contains(sqlType, "JSONB"):
		return TypeString
	case strings.Contains(sqlType, "UUID"), strings.Contains(sqlType, "GUID"):
		return TypeString

	// Tipos booleanos
	case strings.Contains(sqlType, "BOOL"), strings.Contains(sqlType, "BOOLEAN"),
		sqlType == "BIT":
		return TypeBool

	// Tipos de fecha/hora
	case strings.Contains(sqlType, "TIMESTAMP"), strings.Contains(sqlType, "DATETIME"):
		return TypeTime
	case strings.Contains(sqlType, "DATE"), strings.Contains(sqlType, "TIME"):
		return TypeTime
	case strings.Contains(sqlType, "YEAR"):
		return TypeTime

	// Tipos de punto flotante
	case strings.Contains(sqlType, "FLOAT"), strings.Contains(sqlType, "REAL"):
		return TypeFloat
	case strings.Contains(sqlType, "DOUBLE"), strings.Contains(sqlType, "NUMERIC"),
		strings.Contains(sqlType, "DECIMAL"), strings.Contains(sqlType, "DEC"),
		strings.Contains(sqlType, "MONEY"):
		return TypeFloat

	// Tipos binarios (mapear a string para simplificar)
	case strings.Contains(sqlType, "BLOB"), strings.Contains(sqlType, "BINARY"),
		strings.Contains(sqlType, "VARBINARY"), strings.Contains(sqlType, "BYTEA"):
		return TypeString

	default:
		return TypeUnknown
	}
}
