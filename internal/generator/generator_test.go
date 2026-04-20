package generator

import (
	"testing"

	"github.com/flaviogonzalez/gofast/internal/schema"
)

func TestSingular(t *testing.T) {
	cases := map[string]string{
		"users":      "user",
		"products":   "product",
		"categories": "category",
		"companies":  "company",
		"addresses":  "address",
		"boxes":      "box",
		"class":      "class",
		"address":    "address",
		"status":     "status",
		"data":       "data",
		"order":      "order",
		"order_item": "order_item",
		"":           "",
	}
	for in, want := range cases {
		if got := singular(in); got != want {
			t.Errorf("singular(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIDGoType(t *testing.T) {
	mk := func(raw string, unsigned bool) schema.Table {
		return schema.Table{
			Name: "t",
			Columns: []schema.Column{
				{Name: "id", RawType: raw, Unsigned: unsigned, AutoIncrement: true},
			},
		}
	}
	cases := []struct {
		raw      string
		unsigned bool
		want     string
		parse    string
		bits     int
	}{
		{"BIGINT", true, "uint64", "ParseUint", 64},
		{"BIGINT", false, "int64", "ParseInt", 64},
		{"INT", false, "int32", "ParseInt", 32},
		{"INT", true, "uint32", "ParseUint", 32},
		{"SMALLINT", false, "int16", "ParseInt", 16},
		{"TINYINT", true, "uint8", "ParseUint", 8},
	}
	for _, c := range cases {
		tbl := mk(c.raw, c.unsigned)
		if got := idGoType(tbl); got != c.want {
			t.Errorf("idGoType(%s,unsigned=%v)=%s want %s", c.raw, c.unsigned, got, c.want)
		}
		if got := idParseFunc(tbl); got != c.parse {
			t.Errorf("idParseFunc(%s)=%s want %s", c.raw, got, c.parse)
		}
		if got := idBitSize(tbl); got != c.bits {
			t.Errorf("idBitSize(%s)=%d want %d", c.raw, got, c.bits)
		}
	}
}

func TestIDGoTypeFallback(t *testing.T) {
	// No id column, no auto-increment — should fall back to int64.
	tbl := schema.Table{Name: "t", Columns: []schema.Column{{Name: "name", RawType: "VARCHAR"}}}
	if got := idGoType(tbl); got != "int64" {
		t.Errorf("fallback idGoType = %s, want int64", got)
	}
}

func TestUpdatableColumnsExcludesAutoIncrement(t *testing.T) {
	tbl := schema.Table{
		Name: "t",
		Columns: []schema.Column{
			{Name: "id", AutoIncrement: true},
			{Name: "name"},
			{Name: "email"},
		},
	}
	got := updatableColumns(tbl)
	if len(got) != 2 || got[0].Name != "name" || got[1].Name != "email" {
		t.Fatalf("got %+v", got)
	}
}

func TestGenerateQueriesQuotesIdentifiers(t *testing.T) {
	tbl := schema.Table{
		Name: "order",
		Columns: []schema.Column{
			{Name: "id", RawType: "BIGINT", AutoIncrement: true},
			{Name: "status", RawType: "VARCHAR"},
		},
	}
	out, err := GenerateQueriesForTable(tbl)
	if err != nil {
		t.Fatal(err)
	}
	// The reserved-word table name and columns must appear backticked.
	for _, sub := range []string{"`order`", "`status`", "`id`"} {
		if !contains(out, sub) {
			t.Errorf("expected %q in output:\n%s", sub, out)
		}
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && stringsIndex(haystack, needle) >= 0
}

func stringsIndex(s, sub string) int {
	n, m := len(s), len(sub)
	for i := 0; i+m <= n; i++ {
		if s[i:i+m] == sub {
			return i
		}
	}
	return -1
}

func TestToPascal(t *testing.T) {
	cases := map[string]string{
		"id":         "ID",
		"user_id":    "UserID",
		"created_at": "CreatedAt",
		"email":      "Email",
		"image_url":  "ImageURL",
		"order_item": "OrderItem",
		"name":       "Name",
		"":           "",
	}
	for in, want := range cases {
		if got := toPascal(in); got != want {
			t.Errorf("toPascal(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestColumnGoType(t *testing.T) {
	cases := []struct {
		col  schema.Column
		want string
	}{
		{schema.Column{Type: schema.TypeInt, RawType: "BIGINT", Unsigned: false, Nullable: false}, "int64"},
		{schema.Column{Type: schema.TypeInt, RawType: "BIGINT", Unsigned: true, Nullable: false}, "uint64"},
		{schema.Column{Type: schema.TypeInt, RawType: "INT", Unsigned: false, Nullable: true}, "*int32"},
		{schema.Column{Type: schema.TypeString, RawType: "VARCHAR", Nullable: true}, "*string"},
		{schema.Column{Type: schema.TypeString, RawType: "VARCHAR", Nullable: false}, "string"},
		{schema.Column{Type: schema.TypeTime, RawType: "DATETIME", Nullable: true}, "*time.Time"},
		{schema.Column{Type: schema.TypeTime, RawType: "DATETIME", Nullable: false}, "time.Time"},
		{schema.Column{Type: schema.TypeBool, RawType: "BOOL", Nullable: true}, "*bool"},
		{schema.Column{Type: schema.TypeFloat, RawType: "FLOAT", Nullable: false}, "float32"},
		{schema.Column{Type: schema.TypeFloat, RawType: "DOUBLE", Nullable: false}, "float64"},
		{schema.Column{Type: schema.TypeFloat, RawType: "DECIMAL", Nullable: true}, "*float64"},
	}
	for _, c := range cases {
		if got := columnGoType(c.col); got != c.want {
			t.Errorf("columnGoType(%s, nullable=%v) = %q, want %q", c.col.RawType, c.col.Nullable, got, c.want)
		}
	}
}

func TestGenerateModelsFile(t *testing.T) {
	tables := []schema.Table{
		{
			Name: "user",
			Columns: []schema.Column{
				{Name: "id", Type: schema.TypeInt, RawType: "BIGINT", Unsigned: true, AutoIncrement: true},
				{Name: "email", Type: schema.TypeString, RawType: "VARCHAR"},
				{Name: "bio", Type: schema.TypeString, RawType: "TEXT", Nullable: true},
				{Name: "created_at", Type: schema.TypeTime, RawType: "DATETIME", Nullable: true},
			},
		},
	}
	out, err := GenerateModelsFile(tables)
	if err != nil {
		t.Fatal(err)
	}
	for _, sub := range []string{
		"type User struct",
		"ID uint64",
		"Email string",
		"Bio *string",
		"CreatedAt *time.Time",
		"CreateUserParams",
		"UpdateUserParams",
		`"time"`,
	} {
		if !contains(out, sub) {
			t.Errorf("expected %q in output:\n%s", sub, out)
		}
	}
}

func TestGenerateDBFile(t *testing.T) {
	out, err := GenerateDBFile()
	if err != nil {
		t.Fatal(err)
	}
	for _, sub := range []string{"DBTX", "Queries", "func New("} {
		if !contains(out, sub) {
			t.Errorf("expected %q in output:\n%s", sub, out)
		}
	}
}

func TestGenerateQueryFuncsForTable(t *testing.T) {
	tbl := schema.Table{
		Name: "order",
		Columns: []schema.Column{
			{Name: "id", Type: schema.TypeInt, RawType: "BIGINT", AutoIncrement: true},
			{Name: "status", Type: schema.TypeString, RawType: "VARCHAR"},
		},
	}
	out, err := GenerateQueryFuncsForTable(tbl)
	if err != nil {
		t.Fatal(err)
	}
	for _, sub := range []string{
		"func (q *Queries) CreateOrder(",
		"func (q *Queries) GetOrderByID(",
		"func (q *Queries) ListOrder(",
		"func (q *Queries) UpdateOrder(",
		"func (q *Queries) DeleteOrder(",
		"`order`",
		"`status`",
	} {
		if !contains(out, sub) {
			t.Errorf("expected %q in output:\n%s", sub, out)
		}
	}
}

func TestGenerateHandlersSnakeCase(t *testing.T) {
	tbl := schema.Table{
		Name: "order_items",
		Columns: []schema.Column{
			{Name: "id", Type: schema.TypeInt, RawType: "INT", AutoIncrement: true},
			{Name: "order_id", Type: schema.TypeInt, RawType: "INT"},
		},
	}
	out, err := GenerateHandlersForTable(tbl, "testmod")
	if err != nil {
		t.Fatal(err)
	}
	// toPascal should produce "OrderItem" (singular of "order_items")
	for _, sub := range []string{
		"OrderItemHandler",
		"NewOrderItemHandler",
		"models.CreateOrderItemParams",
	} {
		if !contains(out, sub) {
			t.Errorf("expected %q in output:\n%s", sub, out)
		}
	}
}
