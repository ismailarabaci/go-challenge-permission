# Go Permissions Management System

[![CI](https://github.com/ismailarabaci/go-challenge-permission/workflows/CI/badge.svg)](https://github.com/ismailarabaci/go-challenge-permission/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/ismailarabaci/go-challenge-permission)](https://goreportcard.com/report/github.com/ismailarabaci/go-challenge-permission)
[![codecov](https://codecov.io/gh/ismailarabaci/go-challenge-permission/branch/master/graph/badge.svg)](https://codecov.io/gh/ismailarabaci/go-challenge-permission)

A robust, production-ready permissions management system written in Go that handles user management, hierarchical group structures, and fine-grained access control.

## Features

### Core Functionality
- **User Management**: Create and manage users with unique identifiers
- **Hierarchical Groups**: Create user groups with unlimited nesting levels
- **Transitive Membership**: Automatic membership propagation through group hierarchies
- **Cycle Detection**: Prevents circular group dependencies
- **Fine-Grained Permissions**: Control access at user-to-user, user-to-group, group-to-user, and group-to-group levels
- **Permission Checking**: Verify access rights with transitive permission resolution

### Technical Highlights
- **Idiomatic Go**: Follows Go best practices and conventions
- **Database-Backed**: MySQL 8.0 with optimized queries
- **Concurrent-Safe**: Uses Go's goroutines with database connection pooling
- **Zero External Dependencies**: Built with Go standard library only (except MySQL driver)
- **Comprehensive Testing**: 24 unit tests + 2 integration tests with 76%+ coverage
- **CI/CD Pipeline**: Automated testing, linting, and formatting checks

## Quick Start

### Prerequisites
- Go 1.20 or later
- Docker (for MySQL)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/ismailarabaci/go-challenge-permission.git
cd go-challenge-permission
```

2. Start MySQL database:
```bash
docker-compose up -d
```

Or manually:
```bash
docker run \
   --detach \
   --rm \
   --name blp-mysql \
   -e MYSQL_ROOT_PASSWORD=my-secret-pw \
   -e MYSQL_DATABASE=blp-coding-challenge \
   -e MYSQL_USER=blp \
   -e MYSQL_PASSWORD=password \
   --volume=$(pwd)/db/initdb:/docker-entrypoint-initdb.d \
   --publish 3306:3306 \
   mysql:8.0
```

3. Run tests:
```bash
go test ./pkg/server/... -v
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "github.com/BLPDigital/go-challenge-permissions/pkg/server"
)

func main() {
    // Create server with explicit dependency injection
    config := server.DefaultConfig()
    db, err := server.OpenDatabase(config)
    if err != nil {
        panic(err)
    }
    repo := server.NewMySQLRepository(db)
    srv := server.New(repo)
    defer srv.Close()
    
    ctx := context.Background()
    
    // Create users
    alice, _ := srv.CreateUser(ctx, "Alice")
    bob, _ := srv.CreateUser(ctx, "Bob")
    
    // Create group
    admins, _ := srv.CreateUserGroup(ctx, "Admins")
    
    // Add user to group
    srv.AddUserToGroup(ctx, alice, admins)
    
    // Grant permission
    srv.AddUserGroupToUserPermission(ctx, admins, bob)
    
    // Check access
    name, err := srv.GetUserNameWithPermissionCheck(ctx, alice, bob)
    if err == nil {
        fmt.Printf("Access granted! User: %s\n", name)
    }
}
```

## Architecture

### Design Principles

This system follows several key architectural decisions:

1. **Synchronous Database Operations**: Uses Go's natural concurrency model with goroutines and connection pooling rather than async patterns
2. **Repository Pattern**: Separates business logic from data access for testability
3. **Error Handling**: Returns errors instead of panicking for better error propagation
4. **Configuration**: Environment-based configuration with sensible defaults

See [`assumptions_and_design_choices.md`](./assumptions_and_design_choices.md) for detailed architectural decisions and rationale.

### Project Structure

```
.
├── pkg/server/              # Core server implementation
│   ├── config.go           # Configuration management
│   ├── errors.go           # Custom error types
│   ├── interface.go        # Public interfaces (Stage1-Stage5)
│   ├── mysql_repository.go # MySQL data access layer
│   ├── repository.go       # Repository interface
│   ├── server.go           # Server implementation
│   ├── server_test.go      # Unit tests (24 tests)
│   └── integration_test.go # Integration tests (2 scenarios)
├── db/initdb/              # Database schema
│   └── db.sql              # MySQL initialization script
└── .github/workflows/      # CI/CD configuration
    └── ci.yml              # GitHub Actions workflow
```

## Testing

### Run Tests Locally

```bash
# Run all tests
go test ./pkg/server/... -v

# Run with coverage
go test ./pkg/server/... -coverprofile=coverage.txt -covermode=atomic

# Run specific stage tests
go test ./pkg/server/... -v -run Test_Stage1
go test ./pkg/server/... -v -run Test_Stage5

# Run integration tests only
go test ./pkg/server/... -v -run Test_Integration

# Run with race detector
go test ./pkg/server/... -v -race
```

### Test Structure

- **Unit Tests** (`server_test.go`): 24 focused tests organized by stages
  - Stage 1: User management (3 tests)
  - Stage 2: Group management and membership (6 tests)
  - Stage 3: Group hierarchies and cycle detection (4 tests)
  - Stage 4: Transitive membership (3 tests)
  - Stage 5: Permissions and access control (8 tests)

- **Integration Tests** (`integration_test.go`): End-to-end HTTP tests
  - Complex permission scenarios with nested groups
  - Transitive group membership with permissions

### CI Pipeline

This project uses GitHub Actions for continuous integration:

- **Automated Testing**: Tests run on every push and pull request
- **Separate Test Jobs**: Unit tests and integration tests run in parallel
- **MySQL Integration**: Runs with MySQL 8.0 service container
- **Code Quality**: Automated linting with golangci-lint
- **Format Checking**: Ensures consistent code formatting with gofmt
- **Coverage Reporting**: Separate coverage tracking for unit and integration tests

The CI pipeline includes four jobs:
- **Unit Tests**: Runs 24 stage-based unit tests with race detector and coverage
- **Integration Tests**: Runs 2 end-to-end HTTP scenarios with race detector and coverage
- **Lint**: Runs golangci-lint with comprehensive linter configuration
- **Format**: Checks code formatting and runs go vet

## Development

### Database Configuration

The server reads database connection string from the `MYSQL_DSN` environment variable:

```bash
export MYSQL_DSN="user:password@tcp(localhost:3306)/database?parseTime=true"
go test ./pkg/server/... -v
```

Default DSN (if not set): `blp:password@tcp(localhost:3306)/blp-coding-challenge`

### Code Quality

Run linting locally:
```bash
golangci-lint run --timeout=5m
```

Check formatting:
```bash
gofmt -s -l .
go vet ./...
```

## API Reference

### Interfaces

The server implements five progressive interfaces:

- **Stage1**: Basic user operations
- **Stage2**: User group operations
- **Stage3**: Group hierarchies with cycle detection
- **Stage4**: Transitive group membership
- **Stage5**: Permissions and access control

See [`pkg/server/interface.go`](./pkg/server/interface.go) for complete interface definitions.

### Error Handling

Custom errors:
- `UserNotFoundError`: User does not exist
- `GroupNotFoundError`: User group does not exist
- `CycleDetectedError`: Operation would create circular group dependency
- `PermissionDeniedError`: Access denied due to insufficient permissions


## Documentation

- [`challenge_assignment.md`](./challenge_assignment.md) - Original challenge requirements
- [`assumptions_and_design_choices.md`](./assumptions_and_design_choices.md) - Architectural decisions and design rationale
