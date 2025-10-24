# Turbo DB Client Example

This is a simple example application that demonstrates how to use Turbo DB from a Go application.

## What This Example Does

1. Creates a Turbo DB client
2. Creates a new database
3. Lists all databases
4. Gets database details
5. Shows how to connect to the database
6. Cleans up by deleting the database

## Running the Example

### Prerequisites

1. Turbo DB server must be running:
```bash
cd ../..
./bin/turbo-server
```

2. In a new terminal, run this example:
```bash
cd examples/simple-app
go run main.go
```

## Expected Output

```
=== Turbo DB Client Example ===

1. Creating a new database...
✓ Created database: example-app-db (ID: abc-123-def)
  HTTP URL: http://localhost:9000
  Port: 9000
  Status: running

2. Listing all databases...
✓ Found 1 database(s):
  1. example-app-db (Port: 9000, Status: running)

3. Getting database details...
✓ Database details:
  Name: example-app-db
  ID: abc-123-def
  Path: ./databases/example-app-db.db
  Port: 9000
  Status: running
  Created: 2024-10-24T12:00:00Z

4. Connecting to the database...
You can now connect to the database at: http://localhost:9000

Example using libsql-client-go:

	import "github.com/tursodatabase/libsql-client-go/libsql"

	db, err := libsql.NewEmbeddedReplicaDatabase("http://localhost:9000", "")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Execute queries
	conn, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.Execute("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		log.Fatal(err)
	}

5. Waiting for 5 seconds before cleanup...
6. Cleaning up - deleting the database...
✓ Deleted database: example-app-db

=== Example Complete ===
```

## Using as a Library

You can copy the `TurboDBClient` struct and methods into your own application to interact with Turbo DB programmatically.

## API Methods

### Create Database
```go
db, err := client.CreateDatabase("my-database")
```

### List Databases
```go
databases, err := client.ListDatabases()
```

### Get Database
```go
db, err := client.GetDatabase("database-id")
```

### Delete Database
```go
err := client.DeleteDatabase("database-id")
```

## Integration with libSQL

Once you have a database created in Turbo DB, you can connect to it using any libSQL client library:

### Go
```go
import "github.com/tursodatabase/libsql-client-go/libsql"

db, err := libsql.NewEmbeddedReplicaDatabase(dbURL, "")
```

### JavaScript/TypeScript
```javascript
import { createClient } from "@libsql/client";

const client = createClient({ url: dbURL });
```

### Python
```python
import libsql_experimental as libsql

conn = libsql.connect(dbURL)
```

## Next Steps

- Check out the main Turbo DB README for more features
- Explore the REST API documentation
- Try the CLI tool for manual database management
