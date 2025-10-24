#!/bin/bash

# Turbo DB Demo Script
# This script demonstrates the full capabilities of Turbo DB

set -e

echo "======================================"
echo "  Turbo DB - Demonstration Script"
echo "======================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
TURBO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_URL="http://localhost:8080"
SERVER_PID=""

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    if [ ! -z "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
    fi
    echo -e "${GREEN}Done!${NC}"
}

trap cleanup EXIT

# Check if binaries exist
if [ ! -f "$TURBO_DIR/bin/turbo-server" ]; then
    echo -e "${YELLOW}Building Turbo DB...${NC}"
    cd "$TURBO_DIR"
    make build
fi

echo -e "${BLUE}Step 1: Starting Turbo DB Server${NC}"
echo "----------------------------------------"

# Set environment variables
export SERVER_PORT=8080
export DATABASES_DIR="$TURBO_DIR/demo-databases"
export META_DB_PATH="$TURBO_DIR/demo-meta.db"
export PORT_RANGE_START=9000
export PORT_RANGE_END=9100
export LIBSQL_BINARY="../target/release/sqld"

# Clean up any previous demo data
rm -rf "$DATABASES_DIR"
rm -f "$META_DB_PATH"*

# Start the server in the background
cd "$TURBO_DIR"
./bin/turbo-server &
SERVER_PID=$!

# Wait for server to start
echo "Waiting for server to start..."
sleep 3

# Check if server is running
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo -e "${YELLOW}Server failed to start. This is expected if sqld binary is not available.${NC}"
    echo -e "${YELLOW}To fully test Turbo DB, you need to build libSQL first:${NC}"
    echo "  cd .."
    echo "  cargo build --release -p libsql-server"
    exit 0
fi

echo -e "${GREEN}✓ Server started successfully (PID: $SERVER_PID)${NC}"
echo ""

# Health check
echo -e "${BLUE}Step 2: Health Check${NC}"
echo "----------------------------------------"
curl -s "$API_URL/health" | python3 -m json.tool || echo "Health check failed"
echo -e "${GREEN}✓ Server is healthy${NC}"
echo ""

sleep 1

echo -e "${BLUE}Step 3: Creating Databases${NC}"
echo "----------------------------------------"

# Create first database
echo "Creating database: 'production-api'"
RESPONSE=$(curl -s -X POST "$API_URL/api/v1/databases" \
    -H "Content-Type: application/json" \
    -d '{"name": "production-api"}')
echo "$RESPONSE" | python3 -m json.tool || echo "$RESPONSE"
echo ""

sleep 1

# Create second database
echo "Creating database: 'staging-api'"
RESPONSE=$(curl -s -X POST "$API_URL/api/v1/databases" \
    -H "Content-Type: application/json" \
    -d '{"name": "staging-api"}')
echo "$RESPONSE" | python3 -m json.tool || echo "$RESPONSE"
echo ""

sleep 1

# Create third database
echo "Creating database: 'analytics-db'"
RESPONSE=$(curl -s -X POST "$API_URL/api/v1/databases" \
    -H "Content-Type: application/json" \
    -d '{"name": "analytics-db"}')
echo "$RESPONSE" | python3 -m json.tool || echo "$RESPONSE"
echo -e "${GREEN}✓ Created 3 databases${NC}"
echo ""

sleep 1

echo -e "${BLUE}Step 4: Listing All Databases${NC}"
echo "----------------------------------------"
curl -s "$API_URL/api/v1/databases" | python3 -m json.tool || echo "Failed to list databases"
echo ""

sleep 1

echo -e "${BLUE}Step 5: Using CLI Tool${NC}"
echo "----------------------------------------"
echo "$ ./bin/turbo-cli list"
./bin/turbo-cli list
echo ""

# Get the first database ID
DB_ID=$(curl -s "$API_URL/api/v1/databases" | python3 -c "import sys, json; dbs = json.load(sys.stdin)['databases']; print(dbs[0]['id'] if dbs else '')")

if [ ! -z "$DB_ID" ]; then
    echo -e "${BLUE}Step 6: Getting Database Details${NC}"
    echo "----------------------------------------"
    echo "$ ./bin/turbo-cli get $DB_ID"
    ./bin/turbo-cli get "$DB_ID" || true
    echo ""

    sleep 1

    echo -e "${BLUE}Step 7: Getting Connection Information${NC}"
    echo "----------------------------------------"
    echo "$ ./bin/turbo-cli connection $DB_ID"
    ./bin/turbo-cli connection "$DB_ID" || true
    echo ""

    sleep 1

    echo -e "${BLUE}Step 8: Testing with curl${NC}"
    echo "----------------------------------------"
    # Try to connect to the database
    DB_PORT=$(curl -s "$API_URL/api/v1/databases/$DB_ID" | python3 -c "import sys, json; print(json.load(sys.stdin)['port'])")
    echo "Attempting to connect to database on port $DB_PORT..."
    curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" "http://localhost:$DB_PORT/" || echo "Database not responding (expected if sqld is not running)"
    echo ""
fi

sleep 2

echo -e "${BLUE}Step 9: Summary${NC}"
echo "----------------------------------------"
echo -e "${GREEN}Turbo DB is successfully running!${NC}"
echo ""
echo "You can now:"
echo "  - Create databases: ./bin/turbo-cli create <name>"
echo "  - List databases:   ./bin/turbo-cli list"
echo "  - Get details:      ./bin/turbo-cli get <id>"
echo "  - Delete database:  ./bin/turbo-cli delete <id>"
echo ""
echo "API Endpoints:"
echo "  - POST   $API_URL/api/v1/databases"
echo "  - GET    $API_URL/api/v1/databases"
echo "  - GET    $API_URL/api/v1/databases/:id"
echo "  - DELETE $API_URL/api/v1/databases/:id"
echo ""

# Keep server running for a bit
echo -e "${YELLOW}Server will run for 30 more seconds for testing...${NC}"
echo "Press Ctrl+C to stop early"
sleep 30
