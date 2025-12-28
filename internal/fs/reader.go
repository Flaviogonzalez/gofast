package fs

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Database struct {
	Driver   string // "postgres", "mysql", "sqlite"
	Database string // nombre de la DB
}

func ReadDatabaseURL() (string, error) {
	_ = godotenv.Load(".env")

	dbURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dbURL == "" {
		return "", fmt.Errorf("DATABASE_URL no encontrada en las variables de entorno")
	}

	return dbURL, nil
}

func AnalyzeDatabaseURL(dbURL string) (*Database, error) {
	dbURL = strings.TrimSpace(dbURL)
	if dbURL == "" {
		return nil, fmt.Errorf("URL de base de datos vacía")
	}

	u, err := url.Parse(dbURL)
	if err != nil {
		return nil, fmt.Errorf("URL inválida: %w", err)
	}

	db := &Database{}

	switch u.Scheme {
	case "postgres", "postgresql":
		db.Driver = "postgres"
	case "mysql", "mariadb":
		db.Driver = "mysql"
	default:
		return nil, fmt.Errorf("driver no soportado: %s", u.Scheme)
	}

	db.Database = strings.TrimPrefix(u.Path, "/")
	if db.Database == "" {
		return nil, fmt.Errorf("no se especificó el nombre de la base de datos en la URL")
	}

	return db, nil
}
