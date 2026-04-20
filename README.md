# gofast

A rapid REST API generator for Go that scaffolds complete CRUD APIs from MySQL database schemas.

**gofast is a single binary with zero external dependencies.** No Atlas CLI, no sqlc — just Go.

## Prerequisites

- Go 1.23 or later
- MySQL database (local or remote)
- `DATABASE_URL` environment variable pointing to your MySQL database

## Installation

### Using go install (Recommended)

```bash
go install github.com/flaviogonzalez/gofast/cmd/gofast@latest
```

### From Source

```bash
git clone https://github.com/flaviogonzalez/gofast.git
cd gofast
go build -o gofast ./cmd/gofast
```

## Usage

1. Set your database connection string:

**Windows (PowerShell):**
```powershell
$env:DATABASE_URL="mysql://user:password@localhost:3306/mydb"
```

**Linux/Mac:**
```bash
export DATABASE_URL="mysql://user:password@localhost:3306/mydb"
```

2. Create a directory for your project and navigate to it:

```bash
mkdir my-api && cd my-api
```

3. Generate your API:

```bash
gofast generate
```

4. Install dependencies and run:

```bash
go mod tidy
go run cmd/api/main.go
```

Your API starts on port 8080 (or `$PORT` if set).

### Generated Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go          # Application entry point
├── internal/
│   ├── handlers/            # HTTP handlers for each table
│   │   ├── users_handler.go
│   │   └── ...
│   └── models/              # Generated type-safe database models
│       ├── db.go
│       ├── models.go
│       └── <table>_query_funcs.go
├── sql/
│   ├── queries/             # Generated SQL query files
│   └── schema/
│       └── schema.sql       # Inspected database schema
└── go.mod
```

### Generated Endpoints

For each table in your database, gofast generates:

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/{table}` | Create a record |
| `GET` | `/api/{table}` | List all records |
| `GET` | `/api/{table}/{id}` | Get a record by ID |
| `PUT` | `/api/{table}/{id}` | Update a record |
| `DELETE` | `/api/{table}/{id}` | Delete a record |

Plus a health check:
- `GET /health` → `200 OK`

## What gofast Excels At

### Single Binary, Zero Install Friction

gofast ships as a single self-contained binary. It connects directly to your MySQL database using the standard `database/sql` driver — no Atlas CLI, no sqlc, no extra tools to install.

### Zero Configuration

gofast uses your database name as the Go module name. No config files, no interactive prompts.

### Type-Safe Code Generation

All database operations are generated from your schema into typed Go structs and query functions — no reflection, no runtime SQL surprises.

### Production-Ready Patterns

Generated code includes:
- Proper error handling
- HTTP middleware (logging, recovery, request IDs via chi)
- Environment-based configuration
- Clean separation of concerns

### Rapid Prototyping

From database schema to running API in seconds. Perfect for proof of concepts, internal tools, and initial system design.

### MySQL Compatibility

Designed for MySQL with proper handling of `AUTO_INCREMENT`, `LAST_INSERT_ID()`, and MySQL-specific data types.

## Pitfalls and Limitations

### Database Support

Currently only supports MySQL databases. PostgreSQL and SQLite are not yet supported.

### Schema Changes

Regeneration is full project replacement — there is no incremental update support. If you modify your schema, delete the generated project and run `gofast generate` again.

### Complex Queries

Generated queries are basic CRUD only. Joins, aggregations, and business logic must be added manually.

### Authentication and Authorization

No auth is generated. You must implement your own security layer.

### Input Validation

Validation is minimal — JSON is decoded and passed directly to the database. Domain-specific validation must be added manually.

### Table Naming

Singular/plural conversion uses a small set of English rules. Tables with irregular plurals may produce unexpected handler names.

### Primary Key Assumptions

Each table must have an `id` column as its primary key. Tables with composite keys or non-standard primary keys are not yet supported.

## How to Contribute

Contributions are welcome! Here's how you can help:

### Reporting Issues

1. Check if the issue already exists in the GitHub issue tracker
2. Provide a minimal reproduction case
3. Include your Go version, database version, and operating system
4. Share relevant error messages and logs

### Submitting Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Test your changes thoroughly
5. Commit with clear messages (`git commit -m 'Add support for PostgreSQL'`)
6. Push to your branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Areas for Contribution

**High Priority:**
- PostgreSQL support
- SQLite support
- Better plural/singular handling
- Support for composite primary keys
- Incremental schema updates (re-generate without full project deletion)

## License

MIT License - see LICENSE file for details
