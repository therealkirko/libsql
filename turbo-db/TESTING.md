# Testing Guide for Turbo DB

This guide provides comprehensive testing instructions for Turbo DB.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Building](#building)
3. [Manual Testing](#manual-testing)
4. [API Testing](#api-testing)
5. [CLI Testing](#cli-testing)
6. [Integration Testing](#integration-testing)
7. [Performance Testing](#performance-testing)
8. [Troubleshooting](#troubleshooting)

## Prerequisites

Before testing Turbo DB, ensure you have:

1. **Go 1.20+** installed
2. **libSQL built** with sqld binary available:
   ```bash
   cd /path/to/libsql
   cargo build --release -p libsql-server
   # This creates: target/release/sqld
   ```
3. **curl** for API testing
4. **jq** (optional) for JSON formatting

## Building

Build Turbo DB components:

```bash
cd turbo-db

# Build everything
make build

# Or build individually
make build-server   # Builds bin/turbo-server
make build-cli      # Builds bin/turbo-cli
```

Verify the binaries:
```bash
./bin/turbo-server --help  # Should fail gracefully (no --help flag)
./bin/turbo-cli --help     # Should show CLI help
```

## Manual Testing

### 1. Start the Server

```bash
# From turbo-db directory
export LIBSQL_BINARY=../target/release/sqld
export DATABASES_DIR=./test-databases
export META_DB_PATH=./test-meta.db
export PORT_RANGE_START=9000
export PORT_RANGE_END=9100

./bin/turbo-server
```

Expected output:
```
2024/10/24 12:00:00 Starting Turbo DB Server on 0.0.0.0:8080
2024/10/24 12:00:00 Database directory: ./test-databases
2024/10/24 12:00:00 Port range: 9000-9100
2024/10/24 12:00:00 Metadata store initialized
2024/10/24 12:00:00 Database manager initialized
2024/10/24 12:00:00 Server listening on 0.0.0.0:8080
```

### 2. Health Check

In a new terminal:
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "service": "turbo-db"
}
```

## API Testing

### Test 1: Create a Database

```bash
curl -X POST http://localhost:8080/api/v1/databases \
  -H "Content-Type: application/json" \
  -d '{"name": "test-db-1"}' | jq
```

Expected response (200):
```json
{
  "database": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "test-db-1",
    "path": "./test-databases/test-db-1.db",
    "port": 9000,
    "status": "running",
    "created_at": "2024-10-24T12:00:00Z",
    "updated_at": "2024-10-24T12:00:00Z",
    "process_id": 12345,
    "http_url": "http://localhost:9000"
  },
  "message": "Database created successfully"
}
```

Verify the database process is running:
```bash
ps aux | grep sqld
# Should show the sqld process
```

### Test 2: List Databases

```bash
curl http://localhost:8080/api/v1/databases | jq
```

Expected response:
```json
{
  "databases": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "test-db-1",
      "path": "./test-databases/test-db-1.db",
      "port": 9000,
      "status": "running",
      ...
    }
  ],
  "count": 1
}
```

### Test 3: Get Database Details

```bash
DB_ID="550e8400-e29b-41d4-a716-446655440000"  # Replace with actual ID
curl http://localhost:8080/api/v1/databases/$DB_ID | jq
```

### Test 4: Get Connection Info

```bash
curl http://localhost:8080/api/v1/databases/$DB_ID/connection | jq
```

Expected response:
```json
{
  "database_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "test-db-1",
  "http_url": "http://localhost:9000",
  "ws_url": "ws://localhost:9000",
  "status": "running"
}
```

### Test 5: Test Database Connection

Connect to the database directly:
```bash
curl http://localhost:9000/
# Should get a response from sqld
```

### Test 6: Stop Database

```bash
curl -X POST http://localhost:8080/api/v1/databases/$DB_ID/stop | jq
```

Expected response:
```json
{
  "message": "Database stopped successfully"
}
```

Verify the process is stopped:
```bash
ps aux | grep sqld
# The process for this database should be gone
```

### Test 7: Delete Database

```bash
curl -X DELETE http://localhost:8080/api/v1/databases/$DB_ID | jq
```

Expected response:
```json
{
  "message": "Database deleted successfully"
}
```

Verify the database file is removed:
```bash
ls ./test-databases/
# test-db-1.db should be gone
```

### Test 8: Error Cases

**Try to create duplicate database:**
```bash
curl -X POST http://localhost:8080/api/v1/databases \
  -H "Content-Type: application/json" \
  -d '{"name": "test-db-1"}'

curl -X POST http://localhost:8080/api/v1/databases \
  -H "Content-Type: application/json" \
  -d '{"name": "test-db-1"}'
```

Expected: Second request should return 500 with error message.

**Try to get non-existent database:**
```bash
curl http://localhost:8080/api/v1/databases/invalid-id
```

Expected: 404 error.

## CLI Testing

### Test 1: Create Database

```bash
./bin/turbo-cli create my-cli-db
```

Expected output:
```
✓ Database created successfully!

ID:       550e8400-e29b-41d4-a716-446655440001
Name:     my-cli-db
Status:   running
HTTP URL: http://localhost:9001
Port:     9001
```

### Test 2: List Databases

```bash
./bin/turbo-cli list
```

Expected output:
```
ID        NAME       STATUS   PORT  HTTP URL               CREATED
──        ────       ──────   ────  ────────               ───────
550e8400  my-cli-db  running  9001  http://localhost:9001  2024-10-24 12:00

Total: 1 databases
```

### Test 3: Get Database

```bash
./bin/turbo-cli get <database-id>
```

### Test 4: Get Connection Info

```bash
./bin/turbo-cli connection <database-id>
```

### Test 5: Stop Database

```bash
./bin/turbo-cli stop <database-id>
```

### Test 6: Delete Database

```bash
./bin/turbo-cli delete <database-id>
# Type 'yes' to confirm

# Or skip confirmation:
./bin/turbo-cli delete <database-id> --yes
```

### Test 7: CLI with API Key

```bash
# Start server with auth enabled
export ENABLE_AUTH=true
export API_KEY=my-secret-key
./bin/turbo-server &

# Use CLI with API key
./bin/turbo-cli --api-key my-secret-key list

# Or set environment variable
export TURBO_API_KEY=my-secret-key
./bin/turbo-cli list
```

## Integration Testing

### Test Database Creation and Usage

```bash
# 1. Create a database
DB_RESPONSE=$(./bin/turbo-cli create integration-test)
DB_ID=$(echo "$DB_RESPONSE" | grep "ID:" | awk '{print $2}')
DB_URL=$(echo "$DB_RESPONSE" | grep "HTTP URL:" | awk '{print $3}')

echo "Created database $DB_ID at $DB_URL"

# 2. Wait for it to be ready
sleep 2

# 3. Try to connect
curl -v $DB_URL/

# 4. Clean up
./bin/turbo-cli delete $DB_ID --yes
```

### Test Multiple Databases

```bash
# Create 5 databases
for i in {1..5}; do
  ./bin/turbo-cli create "test-db-$i"
  sleep 1
done

# List them
./bin/turbo-cli list

# Verify they're all running
for i in {1..5}; do
  curl http://localhost:$((9000 + i - 1))/ || echo "DB $i not responding"
done

# Delete all
for id in $(curl -s http://localhost:8080/api/v1/databases | jq -r '.databases[].id'); do
  ./bin/turbo-cli delete $id --yes
done
```

### Test Server Restart

```bash
# Create some databases
./bin/turbo-cli create persistent-db-1
./bin/turbo-cli create persistent-db-2

# Note the IDs and ports
./bin/turbo-cli list

# Stop the server (Ctrl+C or kill the process)

# Restart the server
./bin/turbo-server &

# Wait for it to start
sleep 3

# List databases again
./bin/turbo-cli list

# Verify they're restored and running
```

## Performance Testing

### Test Port Exhaustion

```bash
# Try to create more databases than available ports
# Default range is 9000-9100 (100 ports)

for i in {1..105}; do
  ./bin/turbo-cli create "perf-test-$i" || echo "Failed at $i"
done

# Expected: First 100 should succeed, last 5 should fail
```

### Test Rapid Creation

```bash
# Test creating databases quickly
time for i in {1..10}; do
  ./bin/turbo-cli create "rapid-$i" &
done
wait

./bin/turbo-cli list
```

### Test Database Cleanup

```bash
# Create many databases
for i in {1..50}; do
  ./bin/turbo-cli create "cleanup-test-$i"
done

# Delete them all quickly
time for id in $(curl -s http://localhost:8080/api/v1/databases | jq -r '.databases[].id'); do
  curl -X DELETE http://localhost:8080/api/v1/databases/$id &
done
wait

# Verify all are gone
./bin/turbo-cli list
```

## Automated Test Script

Run the comprehensive demo script:

```bash
./demo.sh
```

This script will:
1. Start the server
2. Create multiple databases
3. Test all API endpoints
4. Test CLI commands
5. Clean up automatically

## Troubleshooting

### Server Won't Start

**Problem:** Server exits immediately

**Check:**
1. Is the libSQL binary path correct?
   ```bash
   echo $LIBSQL_BINARY
   ls -l $LIBSQL_BINARY
   ```

2. Is the port already in use?
   ```bash
   lsof -i :8080
   ```

3. Check server logs for errors

**Solution:**
```bash
# Use a different port
export SERVER_PORT=8081
./bin/turbo-server
```

### Database Creation Fails

**Problem:** Database creates but doesn't start

**Check:**
1. sqld binary permissions:
   ```bash
   ls -l $LIBSQL_BINARY
   # Should be executable
   ```

2. Check database logs:
   ```bash
   cat ./databases/logs/<database-name>.log
   ```

3. Try running sqld manually:
   ```bash
   $LIBSQL_BINARY --help
   ```

### Port Allocation Issues

**Problem:** "No available ports" error

**Solution:**
```bash
# Increase port range
export PORT_RANGE_END=9200
./bin/turbo-server
```

### Cleanup After Testing

```bash
# Stop the server
pkill turbo-server

# Clean up test data
make clean

# Or manually:
rm -rf test-databases/
rm -f test-meta.db*
rm -rf demo-databases/
rm -f demo-meta.db*
```

## Success Criteria

A successful test run should:

- ✅ Server starts without errors
- ✅ Health check returns 200
- ✅ Can create databases via API
- ✅ Can create databases via CLI
- ✅ Created databases are accessible via HTTP
- ✅ Can list all databases
- ✅ Can get individual database details
- ✅ Can stop databases
- ✅ Can delete databases
- ✅ Database processes are properly cleaned up
- ✅ Server shutdown is graceful
- ✅ Databases persist and restore after server restart

## Next Steps

After successful testing:

1. Try the example application in `examples/simple-app/`
2. Build a real application using Turbo DB
3. Deploy to production (see DEPLOYMENT.md)
4. Set up monitoring and alerting
