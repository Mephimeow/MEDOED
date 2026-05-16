# MEDOED EDR Agent

## Build

### Agent (C++)
```bash
cd src/agent
mkdir -p build && cd build
cmake .. && make
```

### Backend (Go)
```bash
cd src/backend
go mod download
go build -o server .
```

## Lint

### Agent
```bash
cppcheck --enable=all src/ include/
```

### Backend
```bash
cd src/backend
go vet ./...
gofmt -s -w .
```

## Run Tests

### Agent
```bash
cd src/agent
mkdir -p build && cd build
cmake .. && make
./agent_tests
```

Or with CTest:
```bash
cd src/agent/build
ctest --output-on-failure
```

### Backend
```bash
cd src/backend
go test -v ./...
```

### Backend with coverage
```bash
cd src/backend
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Docker Build

```bash
docker-compose build
```

## Docker Run

```bash
docker-compose up -d
```

## Test Categories

### Agent Tests
- `logger_test.cpp` - Logger functionality
- `heartbeat_test.cpp` - Heartbeater class
- `event_collector_test.cpp` - EventCollector class
- `json_test.cpp` - JSON escaping, timestamp formatting
- `system_test.cpp` - System info collection, /proc access

### Backend Tests
- `models/models_test.go` - Model serialization
- `handlers/handlers_test.go` - HTTP handlers

## Database

Requires PostgreSQL. Schema is in `src/backend/migrations/01_schema.sql`.

Environment variables for database:
- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - Database user (default: medoed)
- `DB_PASSWORD` - Database password (default: medoed_secret)
- `DB_NAME` - Database name (default: medoed)

## Integration Testing

To run integration tests with database:
```bash
# Start postgres
docker run -d --name medoed-test-postgres \
  -e POSTGRES_USER=medoed \
  -e POSTGRES_PASSWORD=medoed_secret \
  -e POSTGRES_DB=medoed_test \
  -p 5433:5432 \
  postgres:16-alpine

# Run tests
DB_HOST=localhost DB_PORT=5433 go test -v ./...

# Cleanup
docker stop medoed-test-postgres && docker rm medoed-test-postgres
```