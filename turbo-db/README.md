# Turbo DB

A Turso-like database management system built on top of libSQL. Turbo DB allows you to create, manage, and interact with multiple libSQL database instances through a simple REST API.

## Features

- **Database Management**: Create, list, stop, and delete libSQL databases via REST API
- **Process Management**: Automatically spawn and manage libSQL server processes
- **Port Management**: Intelligent port allocation and management
- **Metadata Storage**: Track all databases in a SQLite metadata store
- **CLI Tool**: Command-line interface for easy database management
- **Authentication**: Optional API key authentication
- **Graceful Shutdown**: Properly handles shutdown and cleanup of all database processes

## Architecture

```
┌─────────────────────────────────────────┐
│   Turbo DB API Server (Go)              │
│   - REST API                             │
│   - Database Manager                     │
│   - Process Lifecycle Management         │
└─────────────┬───────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────┐
│   libsql-server (sqld) instances        │
│   - Individual database processes       │
│   - HTTP/WebSocket endpoints            │
│   - SQL query execution                 │
└─────────────────────────────────────────┘
```

## Prerequisites

- Go 1.20 or later
- libSQL (sqld binary) built and available
- Linux/macOS (Windows support coming soon)

## Installation

### 1. Build libSQL

First, build the libSQL server binary:

```bash
cd /path/to/libsql
cargo build --release -p libsql-server
```

This will create the `sqld` binary at `target/release/sqld`.

### 2. Build Turbo DB

```bash
cd turbo-db
go mod tidy
make build
```

This will create two binaries:
- `bin/turbo-server` - The API server
- `bin/turbo-cli` - The CLI tool

## Quick Start

### 1. Start the Server

```bash
# From the turbo-db directory
./bin/turbo-server
```

Or with custom configuration:

```bash
export SERVER_PORT=8080
export DATABASES_DIR=./my-databases
export PORT_RANGE_START=9000
export PORT_RANGE_END=9100
./bin/turbo-server
```

### 2. Create a Database

Using the CLI:

```bash
./bin/turbo-cli create my-first-db
```

Using curl:

```bash
curl -X POST http://localhost:8080/api/v1/databases \
  -H "Content-Type: application/json" \
  -d '{"name": "my-first-db"}'
```

### 3. List Databases

Using the CLI:

```bash
./bin/turbo-cli list
```

Using curl:

```bash
curl http://localhost:8080/api/v1/databases
```

### 4. Connect to Your Database

Each database runs on its own port and can be accessed via HTTP:

```bash
# Get connection info
./bin/turbo-cli connection <database-id>

# Query the database using libSQL client
# The database is accessible at the HTTP URL shown
```

## Configuration

Configuration is done via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | API server port | `8080` |
| `SERVER_HOST` | API server host | `0.0.0.0` |
| `META_DB_PATH` | Path to metadata database | `./turbo-meta.db` |
| `DATABASES_DIR` | Directory for database files | `./databases` |
| `LIBSQL_BINARY` | Path to sqld binary | `../target/release/sqld` |
| `PORT_RANGE_START` | Start of port range for databases | `9000` |
| `PORT_RANGE_END` | End of port range for databases | `9100` |
| `API_KEY` | API key for authentication | (empty) |
| `ENABLE_AUTH` | Enable API key authentication | `false` |

### Example .env File

```env
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
DATABASES_DIR=/var/lib/turbo-db/databases
LIBSQL_BINARY=/usr/local/bin/sqld
PORT_RANGE_START=9000
PORT_RANGE_END=9200
ENABLE_AUTH=true
API_KEY=your-secret-api-key-here
```

## REST API Endpoints

### Create Database
```http
POST /api/v1/databases
Content-Type: application/json

{
  "name": "my-database"
}
```

### List Databases
```http
GET /api/v1/databases
```

### Get Database
```http
GET /api/v1/databases/:id
```

### Delete Database
```http
DELETE /api/v1/databases/:id
```

### Stop Database
```http
POST /api/v1/databases/:id/stop
```

### Get Connection Info
```http
GET /api/v1/databases/:id/connection
```

### Health Check
```http
GET /health
```

## CLI Commands

```bash
# Create a database
turbo-cli create <name>

# List all databases
turbo-cli list

# Get database details
turbo-cli get <id>

# Delete a database
turbo-cli delete <id>

# Stop a database
turbo-cli stop <id>

# Get connection information
turbo-cli connection <id>

# Use custom API URL
turbo-cli --url http://localhost:8080 list

# Use API key for authentication
turbo-cli --api-key your-key-here list
# Or set environment variable
export TURBO_API_KEY=your-key-here
turbo-cli list
```

## Development

### Project Structure

```
turbo-db/
├── cmd/
│   ├── server/          # API server main
│   └── cli/             # CLI tool main
├── internal/
│   ├── api/             # HTTP handlers and routing
│   ├── config/          # Configuration management
│   ├── database/        # Metadata store
│   ├── manager/         # Database lifecycle management
│   └── models/          # Data models
├── pkg/
│   └── client/          # Go client library (future)
├── Makefile
├── go.mod
└── README.md
```

### Building

```bash
# Build both server and CLI
make build

# Build server only
make build-server

# Build CLI only
make build-cli

# Run tests
make test

# Clean build artifacts
make clean
```

## Usage Example

Here's a complete example of creating and using a database:

```bash
# 1. Start the server
./bin/turbo-server &

# 2. Create a database
DB_RESPONSE=$(./bin/turbo-cli create myapp-db)
echo "$DB_RESPONSE"

# 3. Extract the HTTP URL and ID
DB_ID=$(echo "$DB_RESPONSE" | grep "ID:" | awk '{print $2}')
DB_URL=$(echo "$DB_RESPONSE" | grep "HTTP URL:" | awk '{print $3}')

# 4. Connect using libSQL client
# Now you can use the libSQL client libraries to connect to $DB_URL
# For example, in Go:
# db, err := sql.Open("libsql", DB_URL)

# 5. When done, delete the database
./bin/turbo-cli delete $DB_ID --yes
```

## Connecting from Your Application

### Go Example

```go
import (
    "database/sql"
    _ "github.com/tursodatabase/go-libsql"
)

func main() {
    // Get the database URL from Turbo DB API
    dbURL := "http://localhost:9000" // From the database creation response

    db, err := sql.Open("libsql", dbURL)
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // Use the database
    _, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
    if err != nil {
        panic(err)
    }
}
```

### JavaScript/TypeScript Example

```javascript
import { createClient } from "@libsql/client";

// Get the database URL from Turbo DB API
const dbUrl = "http://localhost:9000"; // From the database creation response

const client = createClient({
  url: dbUrl,
});

await client.execute("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)");
```

## Roadmap

- [x] Basic database CRUD operations
- [x] Process management
- [x] REST API
- [x] CLI tool
- [x] API key authentication
- [ ] Database restart functionality
- [ ] Database backup/restore
- [ ] Metrics and monitoring
- [ ] Multi-user support with per-database authentication
- [ ] Database templates
- [ ] Resource limits (CPU, memory, connections)
- [ ] Web UI dashboard
- [ ] Docker support
- [ ] Kubernetes operator
- [ ] Database migration tools

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Built on top of [libSQL](https://github.com/tursodatabase/libsql)
- Inspired by [Turso](https://turso.tech)
