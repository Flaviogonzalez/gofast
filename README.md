# gofast

A rapid REST API generator for Go that scaffolds complete CRUD APIs from MySQL database schemas.

## Prerequisites

- Go 1.23 or later
- MySQL database (local or remote)
- sqlc v1.30.0 or later (install from https://docs.sqlc.dev/en/latest/overview/install.html)
- Atlas CLI (install from https://atlasgo.io/getting-started)
- DATABASE_URL environment variable pointing to your MySQL database

## Installation

### Using go install (Recommended)

```bash
go install github.com/flaviogonzalez/gofast/cmd/gofast@latest
```

This will install the `gofast` binary to your `$GOPATH/bin` directory.

### From Source

```bash
git clone https://github.com/flaviogonzalez/gofast.git
cd gofast
go build ./cmd/gofast
```

This will create a `gofast` executable in the current directory.

## Usage

### Basic Usage

1. Set your database connection string:

#### **Environment Variable**

**Windows PowerShell:**
```powershell
$env:DATABASE_URL="driver://user:password@tcp(host:3306)/database"
```

**Linux/Mac:**
```bash
export DATABASE_URL="driver://user:password@tcp(host:3306)/database"
```



2. Create a directory for your new project and navigate to it:

```bash
mkdir my-api
cd my-api
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

Your API will start on port 8080 (or the port specified in the PORT environment variable).

### Generated Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go          # Application entry point
├── internal/
│   ├── handlers/            # HTTP handlers for each table
│   │   ├── users_handler.go
│   │   ├── posts_handler.go
│   │   └── ...
│   └── models/              # sqlc-generated database models
│       ├── db.go
│       ├── models.go
│       └── querier.go
├── sql/
│   ├── queries/             # Generated SQL queries
│   │   ├── users.sql
│   │   ├── posts.sql
│   │   └── ...
│   └── schema/
│       └── schema.sql       # Database schema
├── go.mod
└── sqlc.yaml               # sqlc configuration
```

### Generated Endpoints

For each table in your database, gofast generates the following endpoints:

- `POST /api/{table}` - Create a new record
- `GET /api/{table}` - List all records
- `GET /api/{table}/{id}` - Get a single record by ID
- `PUT /api/{table}/{id}` - Update a record
- `DELETE /api/{table}/{id}` - Delete a record

Plus a health check endpoint:
- `GET /health` - Returns OK if the service is running

## What gofast Excels At

### Zero Configuration

gofast uses your database name as the Go module name, eliminating the need for configuration files or interactive prompts. Just point it at a database and go.

### Type-Safe Code Generation

By leveraging sqlc, all database operations are type-safe at compile time. No reflection, no runtime errors from SQL typos.

### Production-Ready Patterns

Generated code includes:
- Proper error handling
- HTTP middleware (logging, recovery, request IDs)
- Environment-based configuration
- Clean separation of concerns

### Rapid Prototyping

From database schema to running API in seconds. Perfect for:
- Proof of concepts
- Internal tools
- Microservices
- Admin panels
- Mobile app backends

### MySQL Compatibility

Specifically designed for MySQL databases with proper handling of:
- AUTO_INCREMENT columns
- LAST_INSERT_ID() for created records
- MySQL-specific data types

## Pitfalls and Limitations

### Database Support

Currently only supports MySQL databases. PostgreSQL, SQLite, and other databases are not supported.

### Schema Changes

If you modify your database schema after generation, you need to:
1. Delete the generated project
2. Run `gofast generate` again

There is no incremental update support.

### Complex Queries

Generated queries are basic CRUD operations only. Complex queries with joins, aggregations, or business logic must be added manually.

### Authentication and Authorization

No authentication or authorization is generated. You must implement your own security layer.

### Validation

Input validation is minimal. The generated code will decode JSON and pass it to the database, but domain-specific validation must be added.

### Table Naming Assumptions

The singular/plural conversion is simplistic (removes trailing 's'). Tables with irregular plurals may generate odd handler names.

### Primary Key Assumptions

Generated code assumes each table has an `id` column as the primary key. Tables with composite keys or non-standard primary keys may not work correctly.

### Windows Command Execution

The tool uses Windows-specific command execution (`cmd /C`) for running sqlc. On Linux/Mac, this may need adjustment.

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
- Cross-platform command execution (Linux/Mac compatibility)
- Better plural/singular handling
- Support for composite primary keys

**Medium Priority:**
- Incremental schema updates
- Custom query templates
- Middleware customization
- Configuration file support
- Pagination support

**Nice to Have:**
- Authentication scaffolding
- OpenAPI/Swagger documentation generation
- Docker containerization
- Migration management
- Test generation

### Development Setup

1. Clone the repository
2. Install dependencies: `go mod download`
3. Make changes in the appropriate package
4. Build: `go build ./cmd/gofast`
5. Test against a real MySQL database

### Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Add comments for exported functions
- Keep functions focused and small

### Testing

When adding features, please:
- Test against multiple table schemas
- Verify generated code compiles
- Test the generated API endpoints
- Check error handling paths

## License

MIT License - see LICENSE file for details
