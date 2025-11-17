# Assumptions and Design Choices

## Development Approach and Methodology

### Overview

Completed in approximately **one full work day** (~8 hours) with minimal prior Go or Github Actions experience using **Cursor 2.0.77** (Claude Sonnet 4.5).

### Tools and Workflow

**Development Environment:**
- **Cursor 2.0.77:** Plan mode (architecture decisions), Agent mode (code generation), Ask mode (learning Go idioms)
- **AI Model:** Claude Sonnet 4.5 for code generation, architecture guidance, and technical explanations
- **Learning Resources:** [YouTube Go tutorial](https://www.youtube.com/watch?v=8uiZC0l4Ajw), Go documentation, extensive Ask mode queries

**Development Cycle:**
1. **Planning:** Break down requirements, design architecture, identify technical decisions
2. **Learning:** Query AI about idiomatic Go patterns and language-specific concepts
3. **Implementation:** Generate code structure, implement business logic, write tests
4. **Code Review:** Review every single line, ask questions about unfamiliar patterns, adjust for correctness
5. **Iteration:** Run tests, refine implementation, document decisions

### Key Insights

**Advantages:** Rapid learning curve (zero to production in one day), best practices from day one, comprehensive testing naturally integrated, architecture guidance grounded in Go conventions

**Challenges:** Required validation of all AI suggestions, every line needed careful review and understanding, had to distinguish Go-specific idioms from general patterns, could not blindly accept generated code

### Outcomes

✅ All 5 stages implemented with 24 unit + 2 integration tests (all passing)  
✅ Idiomatic Go following community conventions  
✅ Production-ready with error handling, connection pooling, context support  
✅ Zero external dependencies (except MySQL driver)
✅ CI pipeline with AI code reviews

### Conclusion

AI-assisted development with **critical code review** enabled rapid learning and high-quality implementation. Success came from active engagement—scrutinizing, questioning, and adjusting AI-generated code rather than blind acceptance. AI served as an intelligent pair programmer that could be questioned, guided, and validated through active engagement.

---

## Synchronous vs Asynchronous Database Operations

### Decision
The repository layer uses **synchronous database calls** rather than asynchronous patterns.

### Code Comparison

```go
// ✅ CHOSEN: Synchronous (idiomatic Go)
func (r *MySQLRepository) GetUserByID(ctx context.Context, userID int) (string, error) {
    var name string
    err := r.db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", userID).Scan(&name)
    return name, err
}

// Usage in handler (already in its own goroutine)
func handleGetUser(w http.ResponseWriter, r *http.Request) {
    name, err := repo.GetUserByID(r.Context(), userID)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    json.NewEncoder(w).Encode(name)
}
```

```go
// ❌ REJECTED: Asynchronous (unnecessary complexity in Go)
func (r *MySQLRepository) GetUserByID(ctx context.Context, userID int, callback chan<- Result) {
    go func() {
        var name string
        err := r.db.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", userID).Scan(&name)
        callback <- Result{Name: name, Error: err}
    }()
}

// Usage requires channel management
func handleGetUser(w http.ResponseWriter, r *http.Request) {
    resultChan := make(chan Result)
    repo.GetUserByID(r.Context(), userID, resultChan)
    result := <-resultChan
    if result.Error != nil {
        http.Error(w, result.Error.Error(), 500)
        return
    }
    json.NewEncoder(w).Encode(result.Name)
}
```

### Rationale

Go's concurrency model differs fundamentally from JavaScript or Python:

1. **Built-in Concurrency:** Go's `net/http` spawns a goroutine per request automatically—all requests are already concurrent
2. **Connection Pooling:** The `database/sql` package provides automatic, thread-safe connection pooling
3. **Idiomatic Go:** Synchronous code is the Go community standard for repositories—easier to read, test, and debug

### Conclusion
The synchronous pattern leverages Go's goroutines-per-request and connection pooling for excellent performance without async complexity.

## Package Structure: Single `server` Package vs Separate `repository` Package

### Decision
All code in a single **`pkg/server` package** organized into separate files rather than separate packages.

### Code Comparison

```
✅ CHOSEN: Single package (low ceremony)
pkg/server/
  ├── interface.go         # Stage1-5 interfaces
  ├── server.go            # Server implementation
  ├── repository.go        # Repository interface
  ├── mysql_repository.go  # MySQL implementation
  ├── errors.go            # Custom errors
  ├── config.go            # Configuration
  └── server_test.go       # Tests

// All types can be internal
type MySQLRepository struct {
    db *sql.DB  // No need to export
}

// Repository interface enables testing
type Repository interface {
    CreateUser(ctx context.Context, name string) (int, error)
    // ...
}
```

```
❌ REJECTED: Separate packages (added ceremony)
pkg/
  ├── server/
  │   ├── interface.go       # Stage interfaces
  │   ├── server.go          # Must import repository package
  │   ├── errors.go
  │   └── server_test.go
  └── repository/
      ├── repository.go      # Must export interface
      ├── mysql.go           # Must export MySQLRepository
      ├── errors.go          # Must export errors
      └── mysql_test.go

// Types must be exported (forced by package boundary)
type MySQLRepository struct {
    DB *sql.DB  // Must export for cross-package use
}

// Requires import statement everywhere
import "github.com/BLPDigital/go-challenge-permissions/pkg/repository"
```

### Rationale

**Assignment Requirements:** Challenge requires implementing interfaces in `pkg/server/interface.go`

**Go Best Practices:** Start simple, refactor when needed—avoid premature abstraction; package by feature, not by layer

**Benefits:** Low ceremony (no exported types for internal use), easy navigation, appropriate scope (~500-700 lines), Repository interface enables testing

### Conclusion
Single-package structure is appropriate for this focused coding challenge, balancing simplicity with good engineering practices while following Go idioms.

---

## Constructor Pattern: Enforcing Pure Dependency Injection

### Decision
Single `New(repo Repository) *Server` constructor requiring **explicit dependency injection**—no convenience constructors.

### Code Comparison

```go
// ✅ CHOSEN: Explicit dependency injection (testable, flexible)
func New(repo Repository) *Server {
    return &Server{repo: repo}
}

// Usage: All dependencies explicit
config := server.DefaultConfig()
db, err := server.OpenDatabase(config)
if err != nil {
    return err
}
repo := server.NewMySQLRepository(db)
srv := server.New(repo)
defer srv.Close()

// Testing: Easy to inject mocks
mockRepo := &MockRepository{}
srv := server.New(mockRepo)
```

```go
// ❌ REJECTED: Convenience constructor (hides dependencies)
func NewWithDefaults() (*Server, error) {
    config := DefaultConfig()
    db, err := OpenDatabase(config)
    if err != nil {
        return nil, err
    }
    repo := NewMySQLRepository(db)
    return &Server{repo: repo, db: db}, nil
}

// Usage: Dependencies hidden inside
srv, err := server.NewWithDefaults()
defer srv.Close()

// Testing: Must mock at database level (complex)
// OR: Must add separate constructor just for tests
func NewWithMockDB(mockDB *sql.DB) (*Server, error) {
    // Growing constructor explosion
}
```

### Rationale

**Explicit over implicit:** Dependencies always visible, no hidden resource allocation, clear dependency graph

**Testability first:** Mock injection straightforward, test setup explicit about what it creates

**Production-ready:** Supports custom configuration, different environments, different repository implementations

### Conclusion
Pure dependency injection prioritizes explicitness, testability, and flexibility over convenience. Pattern aligns with Go idioms and makes the dependency graph transparent.

---

## Test Structure and Organization

### Decision
Granular unit tests with table-driven patterns plus integration tests for end-to-end validation.

### Code Comparison

```go
// ✅ CHOSEN: Table-driven tests with subtests (scalable, clear)
func Test_Stage1_CreateUser(t *testing.T) {
    tests := []struct {
        name     string
        userName string
    }{
        {name: "create user Alice", userName: "Alice"},
        {name: "create user Bob", userName: "Bob"},
        {name: "create user with spaces", userName: "John Doe"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := setupTestServer(t)
            defer s.Close()
            
            userID, err := s.CreateUser(context.Background(), tt.userName)
            if err != nil {
                t.Fatalf("CreateUser failed: %v", err)
            }
            if userID <= 0 {
                t.Errorf("Expected positive user ID, got %d", userID)
            }
        })
    }
}
```

```go
// ❌ REJECTED: Individual test functions (repetitive, hard to maintain)
func Test_Stage1_CreateUser_Alice(t *testing.T) {
    s := setupTestServer(t)
    defer s.Close()
    
    userID, err := s.CreateUser(context.Background(), "Alice")
    if err != nil {
        t.Fatalf("CreateUser failed: %v", err)
    }
    if userID <= 0 {
        t.Errorf("Expected positive user ID, got %d", userID)
    }
}

func Test_Stage1_CreateUser_Bob(t *testing.T) {
    s := setupTestServer(t)
    defer s.Close()
    
    userID, err := s.CreateUser(context.Background(), "Bob")
    // ... identical code repeated ...
}

func Test_Stage1_CreateUser_WithSpaces(t *testing.T) {
    // ... more duplication ...
}
```

### Approach

**Unit Tests:** Organized by stage (`Test_Stage{X}_{Description}`), table-driven where applicable  
**Integration Tests:** End-to-end HTTP testing with handlers in test file  
**Go Patterns Used:** Table-driven tests, subtests with `t.Run()`, test helpers with `t.Helper()`, `httptest.Server`

### Test Results
24 unit tests + 2 integration tests (11 subtests) — all passing in ~2 seconds

---

## Continuous Integration Pipeline and Branch Protection

### Decision
GitHub Actions CI pipeline with **branch protection** and **trunk-based development** enforced through automated workflows.

### Configuration

**Core CI Jobs** (run on every push/PR):
1. **Unit Tests:** MySQL 8.0 service container, race detector, Codecov reporting
2. **Integration Tests:** End-to-end HTTP scenarios with race detector
3. **Lint:** golangci-lint with comprehensive linters
4. **Format:** Code style enforcement with `gofmt` and `go vet`

**Additional Workflows:**
- **AI Code Review:** Codium PR Agent for automated code review (security, best practices, Go-specific suggestions)
- **PR Automation:** Auto-labeling by file type and size, PR description validation, reviewer assignment
- **Dependabot Auto-Merge:** Auto-approves/merges minor and patch dependency updates after CI passes

**Branch Protection on `master`:**
- Pull requests required (direct commits blocked)
- 1 approval required (self-approval allowed)
- All CI checks must pass
- Conversations must be resolved
- Force pushes blocked

### Rationale

**Trunk-Based Development:** Short-lived feature branches reduce merge conflicts, maintain clean history through squash merges, enable frequent integration

**Branch Protection:** Prevents accidental commits, enforces code review, ensures quality gates, maintains audit trail

**AI Code Review:** Immediate feedback on PRs, catches common issues early, scales review capacity, educates on best practices (chosen Codium PR Agent for free tier, Go support, no API key required)

**Automation Benefits:** Reduces manual overhead, ensures consistency, catches issues early, faster feedback loops

### Conclusion

Branch protection with automated CI, AI review, and PR automation enforces quality gates while enabling high-velocity trunk-based development. The protected master branch combined with automated workflows prevents errors and maintains code standards without manual gatekeeping.

---

## Database Schema: Separate Tables vs Single Principals Table

### Decision
Separate **`users` and `user_groups` tables** rather than a unified `principals` table with type discriminator.

### Code Comparison

```sql
-- ✅ CHOSEN: Separate tables (type-safe foreign keys)
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE user_groups (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE user_group_members (
    user_id INT NOT NULL,
    user_group_id INT NOT NULL,
    PRIMARY KEY (user_id, user_group_id),
    FOREIGN KEY (user_id) REFERENCES users(id),        -- ✅ Type-safe
    FOREIGN KEY (user_group_id) REFERENCES user_groups(id) -- ✅ Type-safe
);

-- Simple queries, no type filter needed
SELECT name FROM users WHERE id = ?;
SELECT name FROM user_groups WHERE id = ?;
```

```sql
-- ❌ REJECTED: Unified principals table (loses type safety)
CREATE TABLE principals (
    id INT AUTO_INCREMENT PRIMARY KEY,
    type ENUM('user', 'group') NOT NULL,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE user_group_members (
    user_id INT NOT NULL,
    user_group_id INT NOT NULL,
    PRIMARY KEY (user_id, user_group_id),
    FOREIGN KEY (user_id) REFERENCES principals(id),      -- ❌ Could be a group!
    FOREIGN KEY (user_group_id) REFERENCES principals(id) -- ❌ Could be a user!
);

-- Cannot enforce: user_id must be type='user', user_group_id must be type='group'
-- Every query needs type filter
SELECT name FROM principals WHERE id = ? AND type = 'user';
SELECT name FROM principals WHERE id = ? AND type = 'group';
```

### Rationale

**Type Safety:** Separate tables enable foreign key constraints that enforce type safety—`user_group_members` can ONLY reference actual users and groups, preventing invalid relationships.

**Simpler Queries:** No type discriminators needed in queries.

**Database-Level Constraints:** Foreign keys enforce business rules impossible to violate: only users in `user_group_members`, only groups in `user_group_hierarchy`

### Conclusion
Separate tables leverage database constraints for business rule enforcement, simpler queries, and accurate domain modeling. The permissions table correctly uses polymorphism where needed (any source → any target), but this doesn't mean principals themselves should be unified.

---

## Permissions Table: No Foreign Key Constraints

### Decision
Polymorphic pattern with type discriminators (`source_type`, `target_type`) but **no foreign key constraints**.

### Code Comparison

```sql
-- ✅ CHOSEN: Polymorphic single table (flexible, simple queries)
CREATE TABLE permissions (
    source_type ENUM('user', 'group') NOT NULL,
    source_id INT NOT NULL,
    target_type ENUM('user', 'group') NOT NULL,
    target_id INT NOT NULL,
    PRIMARY KEY (source_type, source_id, target_type, target_id),
    INDEX idx_source (source_type, source_id),
    INDEX idx_target (target_type, target_id)
    -- ⚠️ No foreign keys - SQL cannot enforce conditional references
);

-- Single unified query for all permission types
SELECT 1 FROM permissions
WHERE source_type = ? AND source_id = ?
  AND target_type = ? AND target_id = ?;

-- Handles all 4 types in one table:
-- User→User, User→Group, Group→User, Group→Group
```

```sql
-- ❌ REJECTED: Separate tables (schema explosion, complex queries)
CREATE TABLE user_to_user_permissions (
    source_user_id INT NOT NULL,
    target_user_id INT NOT NULL,
    PRIMARY KEY (source_user_id, target_user_id),
    FOREIGN KEY (source_user_id) REFERENCES users(id),   -- ✅ FK works here
    FOREIGN KEY (target_user_id) REFERENCES users(id)    -- ✅ FK works here
);

CREATE TABLE user_to_group_permissions (
    source_user_id INT NOT NULL,
    target_group_id INT NOT NULL,
    PRIMARY KEY (source_user_id, target_group_id),
    FOREIGN KEY (source_user_id) REFERENCES users(id),
    FOREIGN KEY (target_group_id) REFERENCES user_groups(id)
);

CREATE TABLE group_to_user_permissions (
    source_group_id INT NOT NULL,
    target_user_id INT NOT NULL,
    PRIMARY KEY (source_group_id, target_user_id),
    FOREIGN KEY (source_group_id) REFERENCES user_groups(id),
    FOREIGN KEY (target_user_id) REFERENCES users(id)
);

CREATE TABLE group_to_group_permissions (
    source_group_id INT NOT NULL,
    target_group_id INT NOT NULL,
    PRIMARY KEY (source_group_id, target_group_id),
    FOREIGN KEY (source_group_id) REFERENCES user_groups(id),
    FOREIGN KEY (target_group_id) REFERENCES user_groups(id)
);

-- Must UNION across 4 tables for permission check
SELECT 1 FROM (
    SELECT 1 FROM user_to_user_permissions WHERE ...
    UNION ALL
    SELECT 1 FROM user_to_group_permissions WHERE ...
    UNION ALL
    SELECT 1 FROM group_to_user_permissions WHERE ...
    UNION ALL
    SELECT 1 FROM group_to_group_permissions WHERE ...
) AS combined_permissions;
```

### Why No Foreign Keys in Chosen Approach

SQL foreign keys **cannot be conditional**—they must reference a single, static table. Cannot write `FOREIGN KEY (source_id) REFERENCES CASE WHEN source_type='user' THEN users(id)...` (this syntax doesn't exist).

### Trade-offs

**What you lose:** No database-level referential integrity, no automatic cascade deletes, orphaned permissions possible

**What you gain:** Simple unified schema (one table handles all 4 types), unified querying, flexible design, no FK overhead

**How it compensates:** Application-level validation through repository interface, controlled API access (no direct SQL), tests ensure entities created before permissions

**Alternatives rejected:** Four separate tables (FKs work but extreme schema complexity), CHECK constraints (not supported in MySQL), triggers (complex, poor performance)

### Conclusion
Deliberate choice driven by polymorphic requirements. SQL cannot enforce conditional FKs. Application-level validation, controlled access, and comprehensive testing compensate for lack of database-enforced referential integrity. Flexibility and simplicity outweigh integrity trade-offs for this use case.

