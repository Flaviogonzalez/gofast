package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempSchema(t *testing.T, sql string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "schema.sql")
	if err := os.WriteFile(p, []byte(sql), 0644); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	return p
}

func TestLoaderParsesBasicTable(t *testing.T) {
	sql := "CREATE TABLE `user` (\n" +
		"  `id` bigint unsigned NOT NULL AUTO_INCREMENT,\n" +
		"  `email` varchar(255) NOT NULL,\n" +
		"  `bio` text NULL,\n" +
		"  PRIMARY KEY (`id`)\n" +
		") ENGINE=InnoDB;\n"
	s, err := NewSQLLoader(writeTempSchema(t, sql)).Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(s.Tables) != 1 {
		t.Fatalf("want 1 table, got %d", len(s.Tables))
	}
	tbl := s.Tables[0]
	if tbl.Name != "user" {
		t.Errorf("name = %q", tbl.Name)
	}
	if len(tbl.Columns) != 3 {
		t.Fatalf("cols = %d", len(tbl.Columns))
	}

	id := tbl.Columns[0]
	if id.Name != "id" || !id.AutoIncrement || !id.Unsigned || id.Nullable || id.RawType != "BIGINT" {
		t.Errorf("bad id col: %+v", id)
	}
	if tbl.Columns[1].Nullable {
		t.Errorf("email should be NOT NULL")
	}
	if !tbl.Columns[2].Nullable {
		t.Errorf("bio should be nullable")
	}
}

func TestLoaderSkipsConstraintsAndReservedNames(t *testing.T) {
	sql := "CREATE TABLE `order` (\n" +
		"  `id` int NOT NULL AUTO_INCREMENT,\n" +
		"  `user_id` int NOT NULL,\n" +
		"  PRIMARY KEY (`id`),\n" +
		"  UNIQUE KEY `uq_user` (`user_id`),\n" +
		"  CONSTRAINT `fk_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)\n" +
		");\n"
	s, err := NewSQLLoader(writeTempSchema(t, sql)).Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(s.Tables) != 1 || s.Tables[0].Name != "order" {
		t.Fatalf("bad tables: %+v", s.Tables)
	}
	if len(s.Tables[0].Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(s.Tables[0].Columns))
	}
}

func TestLoaderParsesMultipleTables(t *testing.T) {
	// Mirrors Atlas's multi-line output format.
	sql := "CREATE TABLE `a` (\n" +
		"  `id` int NOT NULL\n" +
		");\n" +
		"CREATE TABLE `b` (\n" +
		"  `id` int NOT NULL\n" +
		");\n"
	s, err := NewSQLLoader(writeTempSchema(t, sql)).Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(s.Tables) != 2 {
		t.Fatalf("want 2 tables, got %d", len(s.Tables))
	}
}
