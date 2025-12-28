package schema

type Loader interface {
	Load() (*Schema, error)
}

type Schema struct {
	Tables []Table
}

type Table struct {
	Name    string
	Columns []Column
}

type Column struct {
	Name          string
	Type          ColumnType
	Nullable      bool
	AutoIncrement bool
}

type ColumnType string

const (
	TypeInt     ColumnType = "int"
	TypeString  ColumnType = "string"
	TypeBool    ColumnType = "bool"
	TypeTime    ColumnType = "time"
	TypeFloat   ColumnType = "float"
	TypeUnknown ColumnType = "unknown"
)

type ForeignKey struct {
	ColumnName       string
	ReferencedTable  string
	ReferencedColumn string
}
