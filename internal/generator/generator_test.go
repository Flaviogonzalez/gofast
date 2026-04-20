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
