# Go Idioms and Architecture Refactoring Guide

This document outlines the non-idiomatic patterns found in the codebase and explains how to improve them following Go best practices.

---

## 1. Constructor Pattern - Panic vs Error Return

### ❌ Current (Non-Idiomatic)

```go
func New() *server {
    db, err := sql.Open("mysql", "blp:password@tcp(localhost:3306)/blp-coding-challenge")
    if err != nil {
        panic(fmt.Sprintf("sql.Open: %s", err.Error()))
    }

    return &server{
        db: db,
    }
}
```

**Problems:**
- Constructor panics instead of returning an error
- Hardcoded database credentials
- No dependency injection
- Database connection configuration not set
- Connection never closed (resource leak)

### ✅ Improved (Idiomatic)

```go
// Option 1: Accept configured database connection (preferred for testing)
func New(db *sql.DB, repo Repository) *Server {
    return &Server{
        repo: repo,
    }
}

// Option 2: Accept configuration and return error
func NewWithConfig(cfg Config) (*Server, error) {
    db, err := sql.Open("mysql", cfg.DatabaseDSN)
    if err != nil {
        return nil, fmt.Errorf("open database: %w", err)
    }
    
    // Configure connection pool
    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
    
    // Verify connection
    if err := db.PingContext(context.Background()); err != nil {
        db.Close()
        return nil, fmt.Errorf("ping database: %w", err)
    }
    
    repo := NewMySQLRepository(db)
    return New(db, repo), nil
}

// Add Close method for cleanup
func (s *Server) Close() error {
    return s.db.Close()
}
```

**Benefits:**
- No panic - errors are handled by caller
- Dependency injection enables testing with mocks
- Proper connection pool configuration
- Resource cleanup with Close method
- Configuration externalized

---

## 2. Error Handling - Direct Comparison vs errors.Is()

### ❌ Current (Non-Idiomatic)

```go
func (s *server) GetUserName(ctx context.Context, userID int) (string, error) {
    var name string
    err := s.db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", userID).Scan(&name)
    if err != nil {
        if err == sql.ErrNoRows {  // Direct comparison
            return "", fmt.Errorf("user not found: %d", userID)
        }
        return "", fmt.Errorf("failed to get user name: %w", err)
    }
    return name, nil
}
```

**Problems:**
- Direct error comparison `err == sql.ErrNoRows`
- Generic error messages without domain-specific errors
- No custom error types for better error handling

### ✅ Improved (Idiomatic)

```go
// Define custom errors
var (
    ErrUserNotFound = errors.New("user not found")
    ErrGroupNotFound = errors.New("user group not found")
    ErrPermissionDenied = errors.New("permission denied")
    ErrCycleDetected = errors.New("cycle detected in group hierarchy")
)

func (r *MySQLRepository) GetUserByID(ctx context.Context, userID int) (string, error) {
    var name string
    err := r.db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", userID).Scan(&name)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {  // Use errors.Is()
            return "", fmt.Errorf("%w: id=%d", ErrUserNotFound, userID)
        }
        return "", fmt.Errorf("query user: %w", err)
    }
    return name, nil
}

// Caller can check for specific errors
name, err := repo.GetUserByID(ctx, userID)
if errors.Is(err, ErrUserNotFound) {
    // Handle not found case
}
```

**Benefits:**
- `errors.Is()` works with wrapped errors
- Custom error types enable better error handling
- Clients can check for specific error conditions
- Better error context with wrapped errors

---

## 3. Type Visibility - Unexported Server Type

### ❌ Current (Non-Idiomatic)

```go
type server struct {  // Unexported
    db *sql.DB
}

func New() *server {  // Returns unexported type
    // ...
}
```

**Problems:**
- Unexported type limits extensibility
- Inconsistent with Go conventions for public API
- Cannot document type properly in godoc

### ✅ Improved (Idiomatic)

```go
// Server manages user and permission operations.
// It provides a clean business logic layer over the repository.
type Server struct {
    repo Repository
}

// New creates a new Server with the given repository.
func New(repo Repository) *Server {
    return &Server{
        repo: repo,
    }
}
```

**Benefits:**
- Exported type can be extended if needed
- Better godoc documentation
- Follows Go convention for public APIs
- Field is unexported but type is public

---

## 4. Mixed Concerns - Business Logic + Data Access

### ❌ Current (Non-Idiomatic)

```go
func (s *server) CreateUser(ctx context.Context, name string) (int, error) {
    // SQL query directly in business logic
    result, err := s.db.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", name)
    if err != nil {
        return 0, fmt.Errorf("failed to create user: %w", err)
    }
    
    id, err := result.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("failed to get last insert id: %w", err)
    }
    
    return int(id), nil
}
```

**Problems:**
- Database queries mixed with business logic
- Hard to test without a real database
- Cannot swap database implementations
- SQL queries scattered throughout code

### ✅ Improved (Idiomatic)

```go
// Repository interface (pkg/server/repository.go)
type Repository interface {
    CreateUser(ctx context.Context, name string) (int, error)
    GetUserByID(ctx context.Context, userID int) (string, error)
    // ... other methods
}

// MySQL implementation (pkg/server/mysql_repository.go)
type MySQLRepository struct {
    db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
    return &MySQLRepository{db: db}
}

func (r *MySQLRepository) CreateUser(ctx context.Context, name string) (int, error) {
    result, err := r.db.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", name)
    if err != nil {
        return 0, fmt.Errorf("insert user: %w", err)
    }
    
    id, err := result.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("get last insert id: %w", err)
    }
    
    return int(id), nil
}

// Server only has business logic (pkg/server/server.go)
func (s *Server) CreateUser(ctx context.Context, name string) (int, error) {
    // Could add validation, logging, metrics, etc. here
    return s.repo.CreateUser(ctx, name)
}
```

**Benefits:**
- Separation of concerns (business vs data access)
- Easy to test with mock repository
- Can swap MySQL for PostgreSQL, SQLite, etc.
- Single Responsibility Principle

---

## 5. Slice Initialization - Nil Check Pattern

### ❌ Current (Non-Idiomatic)

```go
func (s *server) GetUsersInGroup(ctx context.Context, userGroupID int) ([]int, error) {
    // ... query code ...
    
    var userIDs []int
    for rows.Next() {
        var userID int
        if err := rows.Scan(&userID); err != nil {
            return nil, fmt.Errorf("failed to scan user id: %w", err)
        }
        userIDs = append(userIDs, userID)
    }
    
    // Unnecessary check and conversion
    if userIDs == nil {
        userIDs = []int{}
    }
    
    return userIDs, nil
}
```

**Problems:**
- Unnecessary nil check
- Extra allocation if no rows
- Verbose pattern

### ✅ Improved (Idiomatic)

```go
func (r *MySQLRepository) GetUsersInGroup(ctx context.Context, groupID int) ([]int, error) {
    rows, err := r.db.QueryContext(ctx, 
        "SELECT user_id FROM user_group_members WHERE user_group_id = ? ORDER BY user_id",
        groupID)
    if err != nil {
        return nil, fmt.Errorf("query users in group: %w", err)
    }
    defer rows.Close()
    
    // Pre-allocate with zero capacity to return empty slice if no rows
    userIDs := make([]int, 0)
    
    for rows.Next() {
        var userID int
        if err := rows.Scan(&userID); err != nil {
            return nil, fmt.Errorf("scan user id: %w", err)
        }
        userIDs = append(userIDs, userID)
    }
    
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate rows: %w", err)
    }
    
    return userIDs, nil
}
```

**Benefits:**
- `make([]int, 0)` returns empty slice, never nil
- No nil check needed
- Cleaner, more idiomatic code
- Same behavior with less code

---

## 6. Helper Methods - Visibility and Organization

### ❌ Current (Non-Idiomatic)

```go
// Public method
func (s *server) AddUserGroupToGroup(ctx context.Context, childUserGroupID, parentUserGroupID int) error {
    hasCycle, err := s.wouldCreateCycle(ctx, childUserGroupID, parentUserGroupID)
    // ...
}

// Helper method mixed with public interface methods
func (s *server) wouldCreateCycle(ctx context.Context, childUserGroupID, parentUserGroupID int) (bool, error) {
    // ... implementation
}
```

**Problems:**
- Helper methods mixed with interface methods
- Makes the file harder to navigate
- Helper is specific to implementation detail

### ✅ Improved (Idiomatic)

```go
// In repository layer where it belongs
func (r *MySQLRepository) AddGroupToGroup(ctx context.Context, childID, parentID int) error {
    hasCycle, err := r.wouldCreateCycle(ctx, childID, parentID)
    if err != nil {
        return fmt.Errorf("check cycle: %w", err)
    }
    
    if hasCycle {
        return fmt.Errorf("%w: adding group %d to group %d", ErrCycleDetected, childID, parentID)
    }
    
    _, err = r.db.ExecContext(ctx,
        "INSERT INTO user_group_hierarchy (child_group_id, parent_group_id) VALUES (?, ?)",
        childID, parentID)
    if err != nil {
        return fmt.Errorf("insert hierarchy: %w", err)
    }
    
    return nil
}

// Helper method clearly in repository implementation
func (r *MySQLRepository) wouldCreateCycle(ctx context.Context, childID, parentID int) (bool, error) {
    // ... implementation
}
```

**Benefits:**
- Clear separation of public interface vs implementation
- Helper methods in appropriate layer
- Better organization and discoverability

---

## 7. SQL Query Organization

### ❌ Current (Non-Idiomatic)

```go
func (s *server) GetUsersInGroupTransitive(ctx context.Context, userGroupID int) ([]int, error) {
    // Long SQL query inline
    query := `
        WITH RECURSIVE all_groups AS (
            SELECT ? as group_id
            UNION ALL
            SELECT h.child_group_id
            FROM user_group_hierarchy h
            INNER JOIN all_groups ag ON h.parent_group_id = ag.group_id
        )
        SELECT DISTINCT m.user_id
        FROM user_group_members m
        INNER JOIN all_groups ag ON m.user_group_id = ag.group_id
        ORDER BY m.user_id
    `
    
    rows, err := s.db.QueryContext(ctx, query, userGroupID)
    // ...
}
```

**Problems:**
- SQL queries as local variables
- Hard to find all queries in the codebase
- Not reusable
- Clutters function logic

### ✅ Improved (Idiomatic)

```go
// Define queries as package constants
const (
    queryCreateUser = `INSERT INTO users (name) VALUES (?)`
    queryGetUserByID = `SELECT name FROM users WHERE id = ?`
    
    queryGetUsersTransitive = `
        WITH RECURSIVE all_groups AS (
            SELECT ? as group_id
            UNION ALL
            SELECT h.child_group_id
            FROM user_group_hierarchy h
            INNER JOIN all_groups ag ON h.parent_group_id = ag.group_id
        )
        SELECT DISTINCT m.user_id
        FROM user_group_members m
        INNER JOIN all_groups ag ON m.user_group_id = ag.group_id
        ORDER BY m.user_id`
)

func (r *MySQLRepository) GetUsersInGroupTransitive(ctx context.Context, groupID int) ([]int, error) {
    rows, err := r.db.QueryContext(ctx, queryGetUsersTransitive, groupID)
    if err != nil {
        return nil, fmt.Errorf("query transitive users: %w", err)
    }
    defer rows.Close()
    // ...
}
```

**Benefits:**
- All queries in one place
- Easy to review SQL
- Reusable if needed
- Cleaner function bodies

---

## Summary of Improvements

| Issue | Current | Improved |
|-------|---------|----------|
| Constructor | Panics, hardcoded config | Returns error, accepts config |
| Error handling | Direct comparison `==` | Use `errors.Is()` |
| Error types | Generic strings | Custom sentinel errors |
| Type visibility | Unexported `server` | Exported `Server` |
| Architecture | Mixed concerns | Layered (Server + Repository) |
| Testing | Hard to test | Easy to mock |
| Slice handling | Nil check pattern | Pre-allocate with `make()` |
| SQL queries | Inline strings | Package constants |
| Configuration | Hardcoded | Config struct |

## Implementation Order

1. Create custom error types (`errors.go`)
2. Create configuration struct (`config.go`)
3. Define repository interface (`repository.go`)
4. Implement MySQL repository (`mysql_repository.go`)
5. Refactor server to use repository (`server.go`)
6. Update tests (`server_test.go`)

This refactoring maintains the same functionality while following Go best practices for maintainability, testability, and clarity.

