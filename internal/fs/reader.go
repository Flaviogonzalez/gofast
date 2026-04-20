package fs

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
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

	return encodeCredentials(dbURL), nil
}

// encodeCredentials percent-encodes special characters in the userinfo portion of a database URL
// so that tools like atlas and net/url can parse it correctly.
func encodeCredentials(rawURL string) string {
	// Match scheme://user:pass@host/db
	re := regexp.MustCompile(`^([^:]+)://([^:@]*)(?::([^@]*))?@(.+)$`)
	m := re.FindStringSubmatch(rawURL)
	if m == nil {
		return rawURL
	}
	scheme, user, pass, rest := m[1], m[2], m[3], m[4]
	return scheme + "://" + url.UserPassword(user, pass).String() + "@" + rest
}

// ToMySQLDSN converts a DATABASE_URL in Atlas/URL format
// (e.g. "mysql://user:pass@host:3306/dbname") into a go-sql-driver/mysql DSN
// (e.g. "user:pass@tcp(host:3306)/dbname?parseTime=true").
func ToMySQLDSN(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty database URL")
	}
	if !strings.Contains(raw, "://") {
		return ensureParseTime(raw), nil
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	userInfo := ""
	if u.User != nil {
		if pw, ok := u.User.Password(); ok {
			userInfo = u.User.Username() + ":" + pw
		} else {
			userInfo = u.User.Username()
		}
	}

	host := u.Host
	if host == "" {
		host = "localhost:3306"
	}
	db := strings.TrimPrefix(u.Path, "/")

	dsn := ""
	if userInfo != "" {
		dsn = userInfo + "@"
	}
	dsn += "tcp(" + host + ")/" + db
	if u.RawQuery != "" {
		dsn += "?" + u.RawQuery
	}
	return ensureParseTime(dsn), nil
}

func ensureParseTime(dsn string) string {
	if strings.Contains(dsn, "parseTime=") {
		return dsn
	}
	if strings.Contains(dsn, "?") {
		return dsn + "&parseTime=true"
	}
	return dsn + "?parseTime=true"
}

func AnalyzeDatabaseURL(dbURL string) (*Database, error) {
	dbURL = strings.TrimSpace(dbURL)
	if dbURL == "" {
		return nil, fmt.Errorf("URL de base de datos vacía")
	}

	// Use regex to extract scheme and database name to handle special chars in credentials.
	re := regexp.MustCompile(`^([^:]+)://[^/]*/([^?#\s]+)`)
	matches := re.FindStringSubmatch(dbURL)
	if matches == nil {
		return nil, fmt.Errorf("URL inválida: no se pudo extraer scheme y base de datos")
	}

	db := &Database{}
	scheme := matches[1]
	dbName := matches[2]

	switch scheme {
	case "postgres", "postgresql":
		db.Driver = "postgres"
	case "mysql", "mariadb":
		db.Driver = "mysql"
	default:
		return nil, fmt.Errorf("driver no soportado: %s", scheme)
	}

	if dbName == "" {
		return nil, fmt.Errorf("no se especificó el nombre de la base de datos en la URL")
	}
	db.Database = dbName

	return db, nil
}
