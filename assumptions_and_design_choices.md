# Assumptions and Design Choices

## Development Approach and Methodology

### Overview

This coding challenge was completed in approximately **one full work day** (~8 hours) with minimal prior Go experience, leveraging modern AI-assisted development tools to rapidly learn and implement idiomatic Go patterns.

### Tools and Environment

**Primary Development Tool:**
- **Cursor 2.0.77** - AI-powered code editor
- Fully utilized three core modes:
  - **Plan Mode**: High-level task planning and architecture decisions
  - **Agent Mode**: Autonomous code generation and refactoring
  - **Ask Mode**: Learning Go specifics, idioms, and best practices

**AI Model:**
- **Claude Sonnet 4.5** - Language model for code generation, architecture guidance, and technical explanations

### Learning Approach

**Prior Knowledge:**
- Minimal experience with Go before starting this challenge
- No prior knowledge of Go's concurrency patterns, interfaces, or database libraries

**Learning Resources:**
1. **Video Tutorial**: [Learn Go Programming - Golang Tutorial for Beginners](https://www.youtube.com/watch?v=8uiZC0l4Ajw) - Initial orientation to Go syntax and basics
2. **Official Documentation**: Selective reading of Go documentation for specific topics
3. **AI-Assisted Learning**: Extensive use of Cursor's Ask mode to understand:
   - Go idioms and conventions
   - Package structure and organization
   - Error handling patterns
   - Interface design
   - Testing best practices
   - Concurrency patterns (goroutines, channels)

### Development Workflow

**Typical Development Cycle:**

1. **Planning** (Plan Mode):
   - Break down requirements from challenge stages
   - Design architecture and component boundaries
   - Identify technical decisions to be made

2. **Learning** (Ask Mode):
   - Query about idiomatic Go patterns for specific features
   - Understand Go-specific concepts (e.g., "How do Go interfaces work?")
   - Learn testing patterns and table-driven tests
   - Research database patterns in Go

3. **Implementation** (Agent Mode):
   - Generate initial code structure
   - Implement business logic
   - Write comprehensive tests
   - Refactor for clarity and performance

4. **Code Review and Understanding**:
   - Review every single line of AI-generated code
   - Ask questions about unfamiliar patterns or approaches
   - Make adjustments for clarity or correctness
   - Verify that the code matches intent and requirements
   - Ensure full understanding before moving forward

5. **Iteration**:
   - Run tests and identify failures
   - Use Ask mode to understand errors
   - Refine implementation with Agent mode
   - Document design decisions

### Key Insights from AI-Assisted Development

**Advantages:**

1. **Rapid Learning Curve**
   - From zero Go knowledge to production-quality code in one day
   - AI explains not just "how" but "why" for Go idioms
   - Immediate feedback on design decisions

2. **Best Practices from Day One**
   - No need to discover patterns through trial and error
   - Learned idiomatic Go patterns during implementation
   - Avoided common beginner mistakes

3. **Comprehensive Testing**
   - AI suggested table-driven tests (standard Go pattern)
   - Wrote 24 unit tests + 2 integration tests alongside implementation
   - Achieved high test coverage naturally

4. **Architecture Guidance**
   - AI helped navigate trade-offs (e.g., sync vs async, package structure)
   - Received explanations grounded in Go community conventions
   - Made informed decisions with rationale documented

**Challenges:**

1. **Verification Needed**
   - AI suggestions required validation against actual Go behavior
   - Sometimes received overly complex solutions that needed simplification
   - Had to distinguish between general software principles and Go-specific idioms
   - Every line of generated code required careful review and understanding

2. **Context Switching**
   - Moving between Plan, Agent, and Ask modes required conscious workflow management
   - Balancing autonomous generation with learning and understanding

3. **Active Review Requirement**
   - Cannot blindly accept AI-generated code
   - Must invest time to understand each implementation decision
   - Required asking clarifying questions about unfamiliar patterns
   - Made adjustments when AI solutions didn't align with requirements or best practices


### Code Quality Outcomes

Despite minimal Go experience and rapid development:

- ✅ **All requirements met**: Successfully implemented all 5 stages
- ✅ **Comprehensive testing**: 24 unit tests + 2 integration tests, all passing
- ✅ **Idiomatic Go**: Follows Go community conventions and best practices
- ✅ **Production-ready**: Includes error handling, connection pooling, context support
- ✅ **Well-documented**: Clear code structure with extensive design documentation
- ✅ **Zero external dependencies**: Uses only Go standard library (except MySQL driver)

### Conclusion

This challenge demonstrates that AI-assisted development tools like Cursor with Claude Sonnet 4.5 can enable rapid learning and high-quality implementation even in unfamiliar programming languages. The combination of AI-powered code generation, interactive learning through Ask mode, systematic planning through Plan mode, and **critical code review** created a productive workflow that compressed weeks of traditional learning into a single productive day.

**Key Success Factor**: The active review and understanding of every line of AI-generated code was essential. Rather than blindly accepting generated code, each implementation was scrutinized, questioned, and adjusted as needed. This ensured not only working code but also genuine learning and understanding of Go idioms and patterns.

The resulting codebase is production-ready, follows Go best practices, and includes comprehensive testing—outcomes that would traditionally require significant Go experience to achieve. This was possible not because AI wrote perfect code automatically, but because AI served as an intelligent pair programmer that could be questioned, guided, and validated through active engagement.

---

## Synchronous vs Asynchronous Database Operations

### Decision
The repository layer (`MySQLRepository`) uses **synchronous database calls** rather than asynchronous patterns (callbacks, promises, channels, or futures).

### Rationale

#### Go's Concurrency Model is Different
Unlike JavaScript (Node.js), Python, or other languages where async/await patterns are common, Go handles concurrency fundamentally differently:

1. **Built-in Concurrency at the Request Level**
   - Go's `net/http` server automatically spawns a new goroutine for each incoming HTTP request
   - This means all requests are already handled concurrently without any additional async code
   - Goroutines are extremely lightweight (starts with 2KB stack, grows as needed)
   - You can have millions of goroutines without the overhead of OS threads

2. **Database Connection Pooling is Built-in**
   - The `database/sql` package provides automatic connection pooling
   - Multiple goroutines can safely use the same `*sql.DB` instance concurrently
   - The pool manages connections efficiently, blocking when needed and reusing when available
   - No need for manual async management of database connections

3. **Synchronous Code is Idiomatic Go**
   - The Go community convention for repository patterns is synchronous methods
   - Synchronous code is easier to read, write, test, and debug
   - Error handling is straightforward with Go's explicit error returns
   - The `context.Context` parameter provides cancellation and timeout support

### Code Example

```go
// Current synchronous API (recommended)
func (r *MySQLRepository) GetUserByID(ctx context.Context, userID int) (string, error) {
    return r.queryString(ctx, querySelectUser, &UserNotFoundError{UserID: userID}, "failed to get user name", userID)
}

// Usage in handler (already running in its own goroutine)
func handleGetUser(w http.ResponseWriter, r *http.Request) {
    name, err := repo.GetUserByID(r.Context(), userID)
    if err != nil {
        // handle error
    }
    // use name
}
```

### Performance Characteristics

**How Go achieves high performance:**
- Each HTTP request = separate goroutine = true parallelism
- 1000 concurrent requests = 1000 goroutines handling 1000 DB calls in parallel
- The OS and database connection pool handle the actual I/O efficiently
- No callback hell, no promise chains, no event loop complexity

**Comparison to other languages:**
- **Node.js**: Single-threaded, needs async/await to prevent blocking the event loop
- **Python**: GIL makes threading difficult, asyncio needed for concurrency
- **Go**: Multi-threaded by default, goroutines provide true parallelism

### Trade-offs Considered

#### Advantages of Synchronous API (Current Choice)
✅ Simple, readable code  
✅ Easy error handling  
✅ Idiomatic Go  
✅ Context-based cancellation and timeouts  
✅ Already concurrent at the request level  
✅ Easy to test and mock  

#### Disadvantages of Asynchronous API (Not Chosen)
❌ More complex code (channels, goroutines, synchronization)  
❌ Harder to debug  
❌ More opportunities for race conditions and deadlocks  
❌ Not idiomatic for repository patterns in Go  
❌ No significant performance benefit in typical scenarios  

### Implementation Notes

The current implementation uses:
- `context.Context` for cancellation and timeouts
- Synchronous methods that return `(result, error)`
- Helper methods to reduce code duplication (`execInsert`, `queryString`, `queryIDs`, `queryExists`)
- Explicit error handling at each database interaction

### References

- [Effective Go - Concurrency](https://go.dev/doc/effective_go#concurrency)
- [Go database/sql Tutorial](http://go-database-sql.org/)
- [Go Proverbs](https://go-proverbs.github.io/): "Don't communicate by sharing memory, share memory by communicating"

### Conclusion

The synchronous repository pattern is the correct choice for this Go application. The combination of goroutines-per-request, connection pooling, and context-based cancellation provides excellent performance and scalability without the complexity of async patterns common in other languages.

## Package Structure: Single `server` Package vs Separate `repository` Package

### Decision
All server logic and repository code is kept within a **single `pkg/server` package**, organized into separate files rather than separate packages.

### Current Structure

```
pkg/server/
  interface.go         # Assignment-provided Stage interfaces (Stage1-5)
  server.go            # Server implementation (business logic)
  repository.go        # Repository interface definition
  mysql_repository.go  # MySQL repository implementation
  errors.go            # Custom error types
  config.go            # Configuration structs
  server_test.go       # Tests
```

### Rationale

#### Assignment Requirements
- The challenge explicitly requires implementing interfaces in `pkg/server/interface.go`
- The assignment focuses on server logic, not on demonstrating multi-package architecture
- No requirement to separate concerns across multiple packages

#### Go Best Practices for Small Projects
1. **Start Simple, Refactor When Needed**
   - Go proverb: "A little copying is better than a little dependency"
   - Avoid premature abstraction
   - Add complexity only when there's a concrete need

2. **Package by Feature, Not by Layer**
   - Modern Go discourages generic "repository", "service", "model" packages
   - The `server` package represents a cohesive feature: permission management
   - Internal organization via files provides clarity without package overhead

3. **Standard Library Precedent**
   - `net/http` keeps handler, server, client, and transport together
   - `database/sql` combines driver interface and connection logic
   - Large, cohesive packages are idiomatic in Go

#### Practical Benefits of Single Package

✅ **Simplicity**
- Low ceremony: no need for exported types/functions just for internal use
- Easy navigation: related code is colocated
- Reduced cognitive overhead for reviewers

✅ **Appropriate Scope**
- This is a focused coding challenge, not a microservices architecture
- ~500-700 lines of code fits comfortably in one package
- Repository is only used by this server

✅ **Maintainability**
- Clear file separation provides logical boundaries
- Repository interface already enables testing and mocking
- Easy to refactor later if needs change

✅ **Testability**
- `Repository` interface allows dependency injection
- Can mock repository for unit tests
- Can use real MySQL for integration tests

### Alternative Considered: Separate `pkg/repository` Package

A separate repository package structure would look like:

```
pkg/
  server/
    interface.go       # Stage interfaces
    server.go          # Server implementation
    errors.go          # Server-specific errors
    config.go          # Server configuration
    server_test.go     # Server tests
  repository/
    repository.go      # Repository interface
    mysql.go           # MySQL implementation
    errors.go          # Repository-specific errors
    mysql_test.go      # Repository tests
```

**Advantages:**
- ✅ More explicit separation of concerns
- ✅ Better for reusability across multiple packages
- ✅ Clearer domain boundaries enforced by Go's visibility rules

**Disadvantages:**
- ❌ Unnecessary complexity for this scope
- ❌ Forces exporting internal types/errors that don't need to be public
- ❌ More files and imports to navigate
- ❌ Adds ceremony without clear benefit

### When to Reconsider This Decision

Consider separating into multiple packages if:
- 📦 Multiple packages need to use the repository
- 📦 The codebase grows significantly (multiple thousands of lines)
- 📦 Publishing repository as a separate reusable module
- 📦 Building a production system with multiple microservices
- 📦 Team size increases and ownership boundaries are needed

### Code Organization Principles Applied

1. **Separation via Files**
   - `server.go`: Business logic and Stage interface implementations
   - `repository.go`: Data access interface definition
   - `mysql_repository.go`: Concrete MySQL implementation
   - `errors.go`: Domain-specific error types
   - `config.go`: Configuration management

2. **Interface for Abstraction**
   - `Repository` interface provides testability
   - Enables dependency injection
   - Allows different storage implementations

3. **Clear Naming**
   - File names clearly indicate their purpose
   - No ambiguity about where code belongs

### Conclusion

The single-package structure is the appropriate choice for this coding challenge. It balances simplicity with good software engineering practices, follows Go idioms, and meets all assignment requirements. The file-based organization provides logical separation without the overhead of multiple packages, and the Repository interface ensures the code remains testable and maintainable.

---

## Constructor Pattern: Enforcing Pure Dependency Injection

### Decision
The `Server` constructor enforces **pure dependency injection** with no convenience constructors that hide dependency creation.

### Current API

```go
// Only constructor - requires explicit dependencies
func New(repo Repository) *Server

// Factory functions for creating dependencies
func OpenDatabase(config Config) (*sql.DB, error)
func NewMySQLRepository(db *sql.DB) *MySQLRepository
```

### Usage Pattern

```go
// All dependencies must be explicitly created
config := server.DefaultConfig()
db, err := server.OpenDatabase(config)
if err != nil {
    return err
}
repo := server.NewMySQLRepository(db)
srv := server.New(repo)
defer srv.Close()
```

### Rationale

#### Explicit Over Implicit

**Design Philosophy:**
- Dependencies are always visible at the construction site
- No hidden resource allocation or side effects
- Clear dependency graph makes reasoning about the code easier
- Forces developers to think about their dependencies

**Code Clarity:**
```go
// ✅ GOOD: Dependencies are explicit
db, _ := server.OpenDatabase(config)
repo := server.NewMySQLRepository(db)
srv := server.New(repo)

// ❌ AVOIDED: Hidden dependency creation
srv, _ := server.NewDefault()  // What does this create? Database? Connections?
```

#### Testability First

**Benefits for Testing:**
- Mock injection is straightforward and obvious
- No need to work around convenience constructors
- Test setup is explicit about what it's creating
- Easy to test with different repository implementations

**Example:**
```go
// Production
realRepo := server.NewMySQLRepository(db)
srv := server.New(realRepo)

// Testing
mockRepo := &MockRepository{}
srv := server.New(mockRepo)
```

#### Inversion of Control

**Principles Applied:**
1. **Dependency Inversion Principle**: High-level module (Server) depends on abstraction (Repository interface)
2. **Single Responsibility**: Server constructs business logic, not dependencies
3. **Open/Closed Principle**: Open for extension (new repository types) without modifying Server

**Dependency Flow:**
```
Caller creates:     Config → Database → Repository
Caller injects:     Repository → Server
Server uses:        Repository interface (abstraction)
```

#### Production-Ready Design

**Real-World Considerations:**
- Production code often needs custom configuration
- Database connections may come from connection pools
- Different environments may use different repository implementations
- Microservices may receive dependencies via dependency injection frameworks

**Flexibility Examples:**
```go
// Production with custom config
config := server.Config{
    DatabaseDSN: os.Getenv("DB_DSN"),
    MaxOpenConns: 100,
    MaxIdleConns: 10,
}
db, _ := server.OpenDatabase(config)
repo := server.NewMySQLRepository(db)
srv := server.New(repo)

// Testing with mock
mockRepo := &TestRepository{}
srv := server.New(mockRepo)

// Different storage backend
redisRepo := &RedisRepository{}
srv := server.New(redisRepo)
```

### Alternatives Rejected

#### Option 1: Convenience Constructor
```go
func New() (*Server, error) {
    // Hidden: creates database, repository
    config := DefaultConfig()
    db, _ := OpenDatabase(config)
    repo := NewMySQLRepository(db)
    return &Server{repo: repo}, nil
}
```

**Rejected because:**
- ❌ Hides dependency creation
- ❌ Makes testing harder (must work around it)
- ❌ Reduces flexibility (harder to customize)
- ❌ Obscures resource allocation
- ❌ Creates tight coupling to MySQL

#### Option 2: Multiple Constructors
```go
func New(repo Repository) *Server              // DI version
func NewDefault() (*Server, error)              // Convenience version
func NewWithConfig(config Config) (*Server, error)  // Config version
```

**Rejected because:**
- ❌ Confusing API (which one to use?)
- ❌ Invites use of convenience over best practices
- ❌ More surface area to maintain
- ❌ Tests might use different constructors inconsistently
- ❌ Documentation becomes complex

#### Option 3: Builder Pattern
```go
server.NewBuilder().
    WithDatabase(db).
    WithRepository(repo).
    Build()
```

**Rejected because:**
- ❌ Over-engineering for this use case
- ❌ More verbose than simple function
- ❌ No significant benefit over direct injection
- ❌ Not idiomatic for Go

### Design Trade-offs

#### Advantages of Pure DI (Current Choice)
✅ **Explicit dependencies** - Always visible  
✅ **Maximum testability** - Easy mocking  
✅ **Flexibility** - Any repository implementation  
✅ **No hidden side effects** - Clear resource allocation  
✅ **Production-ready** - Supports custom configuration  
✅ **Simplest API** - Single constructor, easy to understand  
✅ **Forces good practices** - Can't hide behind convenience  

#### Disadvantages Accepted
⚠️ **More verbose** - 4 lines instead of 1 for setup  
⚠️ **Requires understanding** - Developer must know the dependency chain  
⚠️ **No "quick start"** - Can't just call `New()` with no args  

**Mitigation:**
- Clear documentation with examples
- Helper functions (like `setupTestServer` in tests) encapsulate the pattern
- Factory functions (`OpenDatabase`, `NewMySQLRepository`) simplify steps
- The verbosity is intentional - it's self-documenting code

### Consistency with Go Idioms

**Standard Library Examples:**
```go
// net/http - explicit listener injection
listener, _ := net.Listen("tcp", ":8080")
http.Serve(listener, handler)

// database/sql - explicit driver and DSN
db, _ := sql.Open("mysql", dsn)

// os - explicit file descriptors
file, _ := os.Open("/path/to/file")
```

Go's standard library consistently prefers explicit dependency management over convenience constructors that hide resource creation.

### When to Reconsider

Consider adding convenience constructors if:
- 📦 This becomes a library with 1000+ external users demanding simplicity
- 📦 The dependency chain grows beyond 5-6 steps
- 📦 User feedback strongly indicates the explicit pattern is a barrier
- 📦 Production telemetry shows developers frequently misconfigure dependencies

**Current Verdict:** The explicit pattern is appropriate and should be maintained.

### Implementation Notes

**Factory Functions Provided:**
- `OpenDatabase(config Config) (*sql.DB, error)` - Creates configured DB connection
- `NewMySQLRepository(db *sql.DB) *MySQLRepository` - Creates repository
- `DefaultConfig() Config` - Provides sensible defaults

**These are building blocks, not convenience shortcuts.** Each has a single responsibility and returns its resource for the caller to manage.

### Testing Pattern

All tests follow the explicit pattern:

```go
func setupTestServer(t *testing.T) *Server {
    t.Helper()
    
    // Explicit dependency creation - no hidden magic
    config := DefaultConfig()
    db, err := OpenDatabase(config)
    if err != nil {
        t.Fatalf("Failed to open database: %v", err)
    }
    
    repo := NewMySQLRepository(db)
    return New(repo)
}
```

This test helper serves as both:
1. A reusable setup function for tests
2. Documentation of the recommended construction pattern

### Conclusion

Enforcing pure dependency injection through a single `New(repo Repository)` constructor is the right choice for this codebase. It prioritizes explicitness, testability, and flexibility over convenience. The pattern aligns with Go idioms, supports production use cases, and makes the dependency graph transparent. While it requires more lines of code at construction sites, this verbosity is intentional and serves as self-documenting code that makes dependencies visible and manageable.

---

## Test Structure and Organization

### Decision: Granular Unit Tests with Table-Driven Patterns

The test suite is organized into focused unit tests with parametrization and a separate integration test suite for end-to-end validation.

### Test Organization

#### Unit Tests (`server_test.go`)

**Naming Convention:** `Test_Stage{X}_{Description}()`
- Examples: `Test_Stage1_CreateUser()`, `Test_Stage5_DirectUserToUserPermission()`
- Preserves the original stage organization for easy reference
- Each test focuses on a single scenario or behavior
- Clear, descriptive names that explain what is being tested

**Structure:**
- **Stage 1:** User operations (create, get, error handling)
- **Stage 2:** User groups (create, membership, duplicates, empty groups)
- **Stage 3:** Hierarchical groups (nesting, cycle detection)
- **Stage 4:** Transitive membership (multi-level hierarchies)
- **Stage 5:** Permissions (4 scenarios + transitive permissions)

#### Integration Tests (`integration_test.go`)

**Purpose:** End-to-end testing via HTTP layer

**Components:**
- HTTP handlers implemented in test file
- Test helper functions with `t.Helper()` marker
- Two comprehensive integration scenarios:
  1. Complex permission scenario with hierarchical groups
  2. Transitive group membership with 3-level hierarchy

### Idiomatic Go Testing Patterns Used

1. **Table-Driven Tests**
   ```go
   tests := []struct {
       name     string
       userName string
   }{
       {name: "create user Alice", userName: "Alice"},
       {name: "create user Bob", userName: "Bob"},
   }
   for _, tt := range tests {
       t.Run(tt.name, func(t *testing.T) {
           // test logic
       })
   }
   ```

2. **Subtests with `t.Run()`**
   - Enables parallel execution
   - Provides better test output organization
   - Allows running specific subtests

3. **Test Helpers with `t.Helper()`**
   - `setupTestServer()` - Creates test server instances
   - `setupHTTPTestServer()` - Creates HTTP test servers
   - HTTP helper functions for common operations
   - Proper error reporting shows the calling line, not helper line

4. **`httptest.Server` for Integration Tests**
   - Standard Go approach for testing HTTP handlers
   - No external server required
   - Fast and isolated

5. **Proper Cleanup with `defer`**
   - All servers and resources cleaned up properly
   - No resource leaks between tests

6. **Clear Arrange-Act-Assert Structure**
   - Setup phase clearly separated
   - Action being tested is obvious
   - Assertions are explicit and descriptive

### Rationale for This Structure

**Benefits:**

1. **Granularity:** Each test covers a specific scenario, making failures easy to identify
2. **Maintainability:** Table-driven tests make adding new cases trivial
3. **Readability:** Descriptive names and focused tests are self-documenting
4. **Reusability:** Helper functions eliminate code duplication
5. **Coverage:** Multiple test cases cover edge cases and error conditions
6. **Integration Validation:** HTTP layer tests verify end-to-end functionality
7. **Parallelization:** Subtests can run in parallel for faster execution

**Design Choices:**

- 26 test functions: 24 unit tests + 2 integration tests (containing 11 subtests)
- Clear separation between unit tests (repository/business logic) and integration tests (HTTP layer)
- Table-driven pattern is idiomatic Go and widely adopted

### Integration Test Approach

**Why HTTP handlers in test file?**
- Tests the repository through a realistic HTTP layer
- No production HTTP code needed yet (keeps codebase focused)
- Easy to extract to production code later if needed
- Demonstrates how the server would be used in a real application

**Test Coverage:**
- Permission checks via HTTP (403 Forbidden for denied access)
- Context-based authentication (X-Context-User-ID header)
- Complex permission scenarios with nested groups
- Both positive (access granted) and negative (access denied) cases

### Running Tests

```bash
# Run all tests
go test ./pkg/server/... -v

# Run specific stage
go test ./pkg/server/... -v -run Test_Stage1

# Run specific test
go test ./pkg/server/... -v -run Test_Stage5_DirectUserToUserPermission

# Run integration tests only
go test ./pkg/server/... -v -run Test_Integration
```

### Test Results

All tests passing: 24 unit tests + 2 integration tests (with 11 subtests)
- Total execution time: ~2 seconds
- All scenarios covered: user operations, groups, hierarchies, transitive membership, permissions
- Integration tests validate full HTTP request/response cycle

---

## Continuous Integration (CI) Pipeline

### Decision: GitHub Actions for Automated Testing

The project uses GitHub Actions for continuous integration to ensure code quality and catch issues early.

### CI Configuration

**Workflow File:** `.github/workflows/ci.yml`

**Jobs:**
1. **Test Job** - Runs on every push/PR
   - Tests against Go 1.21, 1.22, 1.23 (matrix strategy)
   - Spins up MySQL 8.0 service container
   - Runs full test suite with race detector
   - Generates coverage report
   - Uploads to Codecov

2. **Lint Job** - Code quality checks
   - Runs golangci-lint with custom configuration
   - Checks for common issues, bugs, and style violations
   - Configuration in `.golangci.yml`

3. **Format Job** - Code style enforcement
   - Verifies consistent formatting with `gofmt`
   - Runs `go vet` for suspicious constructs

### Features

**Multi-Version Testing**
- Ensures compatibility across Go versions
- Matrix strategy runs tests in parallel
- Catches version-specific issues early

**Service Containers**
- MySQL 8.0 runs as a service
- Health checks ensure database readiness
- Mirrors production environment

**Caching**
- Go module cache speeds up builds
- Build cache reduces compilation time
- Typical CI run: ~2-3 minutes

**Dependency Management**
- Dependabot configured for weekly updates
- Automated security patches
- Keeps dependencies current

### Linting Configuration

**Enabled Linters:**
- `errcheck`: Unchecked errors
- `gosec`: Security issues
- `govet`: Suspicious constructs
- `staticcheck`: Advanced static analysis
- `gofmt` / `goimports`: Formatting
- `ineffassign`: Ineffective assignments
- `dupl`: Code duplication
- And more (see `.golangci.yml`)

**Test Exemptions:**
- Duplication allowed in test files
- Security checks relaxed for tests
- Focus on production code quality

### Coverage Reporting

- Coverage uploaded to Codecov
- Badge displayed in README
- Tracks coverage trends over time
- Helps identify untested code paths

### Rationale

**Why GitHub Actions?**
- ✅ Native GitHub integration
- ✅ Free for public repositories
- ✅ Easy service containers (MySQL)
- ✅ Matrix builds for multi-version testing
- ✅ Extensive marketplace of actions

**Benefits:**
1. **Early Detection**: Catches issues before merge
2. **Quality Assurance**: Automated checks prevent regressions
3. **Confidence**: Green CI gives confidence to merge
4. **Documentation**: CI config documents build/test process
5. **Visibility**: Badges show project health at a glance

**CI Best Practices Applied:**
- Fast feedback (parallel jobs)
- Comprehensive testing (unit + integration)
- Realistic environment (MySQL service)
- Security scanning (gosec)
- Dependency updates (Dependabot)

---

## Database Schema: Separate Tables vs Single Principals Table

### Decision
The database schema uses **separate `users` and `user_groups` tables** rather than a unified `principals` table with a type discriminator.

### Current Schema

```sql
-- Separate tables approach
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE user_groups (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### Alternative Considered: Single Principals Table

A unified table with type discriminator would look like:

```sql
CREATE TABLE principals (
    id INT AUTO_INCREMENT PRIMARY KEY,
    type ENUM('user', 'group') NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### Rationale

#### Type Safety Through Foreign Key Constraints

**Problem with Single Table:**
With a unified `principals` table, foreign keys can't enforce type constraints:

```sql
-- With principals table - NO TYPE SAFETY
CREATE TABLE user_group_members (
    user_principal_id INT,
    group_principal_id INT,
    FOREIGN KEY (user_principal_id) REFERENCES principals(id),   -- Could reference a group!
    FOREIGN KEY (group_principal_id) REFERENCES principals(id)   -- Could reference a user!
)
```

You could accidentally have:
- A group as a member of another group (in the members table meant for users only)
- A user as a parent in the hierarchy table (meant for groups only)

**Separate Tables Provide Type Safety:**

```sql
-- With separate tables - DATABASE-ENFORCED TYPE SAFETY
CREATE TABLE user_group_members (
    user_id INT,
    user_group_id INT,
    FOREIGN KEY (user_id) REFERENCES users(id),              -- Can ONLY be a user
    FOREIGN KEY (user_group_id) REFERENCES user_groups(id)   -- Can ONLY be a group
)
```

The database schema itself prevents invalid relationships. No application-level validation needed.

#### Simpler Queries Without Type Discriminators

**Single Table Complexity:**
Every query requires filtering by type, adding overhead and potential for errors:

```sql
-- Get user by ID (principals table)
SELECT name FROM principals WHERE id = ? AND type = 'user'

-- Get group by ID (principals table)
SELECT name FROM principals WHERE id = ? AND type = 'group'

-- Get users in group (principals table)
SELECT p.id, p.name 
FROM principals p
JOIN user_group_members m ON p.id = m.user_principal_id
WHERE m.group_principal_id = ? AND p.type = 'user'  -- Must remember type filter!
```

**Separate Tables Simplicity:**

```sql
-- Get user by ID (users table)
SELECT name FROM users WHERE id = ?

-- Get group by ID (user_groups table)
SELECT name FROM user_groups WHERE id = ?

-- Get users in group (users table)
SELECT u.id, u.name
FROM users u
JOIN user_group_members m ON u.id = m.user_id
WHERE m.user_group_id = ?  -- Type filter unnecessary
```

Queries are simpler, shorter, and impossible to get wrong by forgetting the type filter.

#### Better Semantic Clarity

**Domain Model Distinction:**
Users and groups are fundamentally different concepts with different relationships:

1. **Users**:
   - Can be members of groups
   - Cannot contain other users
   - Cannot contain groups
   - Represent individuals

2. **Groups**:
   - Can contain users
   - Can contain other groups (nested hierarchies)
   - Cannot be members of themselves
   - Represent collections

**The Stage Interfaces Enforce This:**

```go
type Stage2 interface {
    CreateUser(ctx context.Context, name string) (int, error)
    CreateUserGroup(ctx context.Context, name string) (int, error)
    AddUserToGroup(ctx context.Context, userID, userGroupID int) error
    // Note: No AddGroupToGroup here - that's Stage3
}

type Stage3 interface {
    Stage2
    AddUserGroupToGroup(ctx context.Context, childUserGroupID, parentUserGroupID int) error
    // Note: Different method, different semantics
}
```

The API treats them differently because they ARE different. The schema should reflect this.

#### Enforces Business Rules at Database Level

**Constraint Examples:**

```sql
-- user_group_members: ONLY users can be members
FOREIGN KEY (user_id) REFERENCES users(id)

-- user_group_hierarchy: ONLY groups can be in hierarchies
FOREIGN KEY (child_group_id) REFERENCES user_groups(id)
FOREIGN KEY (parent_group_id) REFERENCES user_groups(id)
CHECK (child_group_id != parent_group_id)
```

These constraints are **impossible to violate** at the database level. With a principals table, you'd need:
- Application-level validation (can fail)
- Database triggers (complex and error-prone)
- CHECK constraints referencing the same table (limited effectiveness)

**Example of What's Prevented:**

```sql
-- With separate tables: This fails at INSERT time
INSERT INTO user_group_hierarchy (child_group_id, parent_group_id)
VALUES (999, 888);  -- ERROR if either ID references a user table

-- With principals table: This succeeds but is semantically wrong
INSERT INTO group_hierarchy (child_principal_id, parent_principal_id)
VALUES (999, 888);  -- No error even if 999 is a user!
```

#### Better Performance and Indexing

**Index Efficiency:**

With separate tables:
```sql
-- Primary key on users(id) - compact, efficient
-- Primary key on user_groups(id) - compact, efficient
```

With principals table:
```sql
-- Every index must include type to be useful
INDEX idx_user_lookups (type, id)  -- Larger, less efficient
```

**Query Optimization:**

- Smaller tables fit better in buffer pools
- Query planner can optimize better without type filters
- Index-only scans more likely with separate tables
- Partition pruning not needed

**Real-World Impact:**

```sql
-- Separate tables: Table scan of users only (potentially smaller)
SELECT * FROM users WHERE name LIKE 'A%'

-- Principals table: Table scan of ALL principals, filter by type
SELECT * FROM principals WHERE type = 'user' AND name LIKE 'A%'
```

#### The Permissions Table Already Handles Polymorphism

The permissions table correctly uses polymorphic associations where needed:

```sql
CREATE TABLE permissions (
    source_type ENUM('user', 'group') NOT NULL,
    source_id INT NOT NULL,
    target_type ENUM('user', 'group') NOT NULL,
    target_id INT NOT NULL,
    PRIMARY KEY (source_type, source_id, target_type, target_id)
)
```

This is appropriate because:
- Permissions ARE polymorphic by design (any source → any target)
- No foreign key constraints needed (permissions can outlive their targets)
- The relationship itself is the entity, not the principals

This doesn't mean the principals themselves should be in one table.

### Alternatives Rejected

#### Option 1: Single Principals Table with Type Discriminator

```sql
CREATE TABLE principals (
    id INT AUTO_INCREMENT PRIMARY KEY,
    type ENUM('user', 'group') NOT NULL,
    name VARCHAR(255) NOT NULL,
    INDEX idx_type (type)
)
```

**Rejected because:**
- ❌ Loss of type safety at database level
- ❌ Every query needs type discriminator
- ❌ Foreign keys can't enforce business rules
- ❌ More complex and error-prone queries
- ❌ Worse query performance
- ❌ Doesn't match the API's semantic distinction
- ❌ Larger indexes due to type column

#### Option 2: Class Table Inheritance

```sql
CREATE TABLE principals (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL
)

CREATE TABLE users (
    id INT PRIMARY KEY,
    FOREIGN KEY (id) REFERENCES principals(id)
    -- user-specific fields here
)

CREATE TABLE user_groups (
    id INT PRIMARY KEY,
    FOREIGN KEY (id) REFERENCES principals(id)
    -- group-specific fields here
)
```

**Rejected because:**
- ❌ Over-engineering for this use case (no type-specific fields)
- ❌ Every query requires joins
- ❌ No benefit since users and groups have identical structure
- ❌ Added complexity without solving any problems
- ❌ Still can't enforce relationship constraints

#### Option 3: Separate Tables for Each Permission Type

```sql
CREATE TABLE user_to_user_permissions (...)
CREATE TABLE user_to_group_permissions (...)
CREATE TABLE group_to_user_permissions (...)
CREATE TABLE group_to_group_permissions (...)
```

**Rejected because:**
- ❌ Too many tables (schema explosion)
- ❌ Permission queries become complex (UNION of 4 tables)
- ❌ Hard to add new permission types
- ❌ Doesn't address the user/group table question

### Trade-offs

#### Advantages of Separate Tables (Current Choice)

✅ **Database-enforced type safety** - Impossible to violate business rules  
✅ **Simpler queries** - No type filters needed  
✅ **Better performance** - Smaller tables, better indexes  
✅ **Clear semantics** - Schema matches domain model  
✅ **Standard pattern** - Well-understood by developers  
✅ **Easier to reason about** - Clear table boundaries  
✅ **Foreign key constraints work properly** - Type safety for relationships  

#### Disadvantages Accepted

⚠️ **More tables** - 2 tables instead of 1 (minimal cost)  
⚠️ **Slight duplication** - Both have id/name structure (acceptable for different entities)  
⚠️ **ID spaces are separate** - User ID 1 ≠ Group ID 1 (this is actually a feature)  

**Mitigation:**
- The duplication is superficial; the entities have different meanings and relationships
- Separate ID spaces prevent confusion and enforce type safety
- 2 tables is not a maintenance burden

### When to Reconsider This Decision

Consider a unified principals table if:

- 📦 Users and groups become truly interchangeable in the domain model
- 📦 The API changes to treat them polymorphically (e.g., a generic `CreatePrincipal`)
- 📦 You need to frequently query "all principals regardless of type"
- 📦 Users and groups start sharing significant behavioral code
- 📦 The number of entity types explodes (10+ types, not just 2)

**None of these apply to this system.** The Stage interfaces explicitly distinguish between users and groups, and their relationship patterns are fundamentally different.

### Implementation Notes

**Current Schema:**
- `users` table: Contains individual user entities
- `user_groups` table: Contains group entities
- `user_group_members`: Maps users to groups (only users can be members)
- `user_group_hierarchy`: Maps groups to groups (only groups can be nested)
- `permissions`: Polymorphic table for access control (correctly uses type discriminators)

**Query Patterns:**
- User operations use `users` table exclusively
- Group operations use `user_groups` table exclusively
- Membership queries join `users` with `user_group_members`
- Hierarchy queries use recursive CTEs on `user_group_hierarchy`
- Permission queries reference both tables as needed via polymorphic associations

### References

- [Database Design: Vertical Table vs Separate Tables](https://stackoverflow.com/questions/3579079/)
- [Polymorphic Associations in Database Design](https://stackoverflow.com/questions/922184/)
- [Single Table Inheritance vs Class Table Inheritance](https://martinfowler.com/eaaCatalog/singleTableInheritance.html)

### Conclusion

The separate tables approach (`users` and `user_groups`) is the correct design for this permission management system. It leverages database constraints to enforce business rules, results in simpler and faster queries, and accurately models the domain where users and groups are distinct concepts with different relationship patterns. While a unified `principals` table might appear simpler at first glance, it sacrifices type safety, query simplicity, and performance for an abstraction that doesn't align with the actual requirements. The current schema provides the right balance of normalization, performance, and semantic clarity.

---

## Permissions Table: No Foreign Key Constraints

### Decision

The `permissions` table uses a **polymorphic pattern** with type discriminators (`source_type`, `target_type`) but **does not enforce referential integrity** through foreign key constraints.

### Current Schema

```sql
CREATE TABLE permissions (
    source_type ENUM('user', 'group') NOT NULL,
    source_id INT NOT NULL,
    target_type ENUM('user', 'group') NOT NULL,
    target_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (source_type, source_id, target_type, target_id),
    INDEX idx_source (source_type, source_id),
    INDEX idx_target (target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### The Problem: Polymorphic References

The permissions table uses a polymorphic association pattern where the same ID column references different tables depending on a type column:

**`source_id` Can Reference:**
```sql
IF source_type = 'user'  → source_id refers to users.id
IF source_type = 'group' → source_id refers to user_groups.id
```

**`target_id` Can Reference:**
```sql
IF target_type = 'user'  → target_id refers to users.id
IF target_type = 'group' → target_id refers to user_groups.id
```

This creates **four possible combinations** for each permission record:
1. User → User (source_type='user', target_type='user')
2. User → Group (source_type='user', target_type='group')
3. Group → User (source_type='group', target_type='user')
4. Group → Group (source_type='group', target_type='group')

### Why Foreign Keys Are Impossible

SQL foreign key constraints **cannot be conditional** or reference different tables based on a column value. You cannot write:

```sql
-- ❌ THIS SYNTAX DOESN'T EXIST IN SQL
FOREIGN KEY (source_id) REFERENCES 
    CASE 
        WHEN source_type = 'user' THEN users(id)
        WHEN source_type = 'group' THEN user_groups(id)
    END
```

Foreign keys must specify a **single, static target table** at schema definition time. They are evaluated at every INSERT/UPDATE and require a fixed reference table.

### Consequences of No Foreign Keys

#### ❌ What You Lose

**1. No Database-Level Referential Integrity**

The database will not prevent invalid references:

```sql
-- This will succeed even if user ID 999999 doesn't exist
INSERT INTO permissions 
VALUES ('user', 999999, 'user', 1, NOW());

-- This will succeed even if the referenced group was deleted
INSERT INTO permissions 
VALUES ('group', 5, 'user', 10, NOW());
```

**2. No Automatic Cascade Operations**

Deleting users or groups doesn't automatically clean up permissions:

```sql
-- Delete a user
DELETE FROM users WHERE id = 42;

-- Orphaned permissions remain in the table
-- These records now reference a non-existent user
SELECT * FROM permissions 
WHERE (source_type = 'user' AND source_id = 42)
   OR (target_type = 'user' AND target_id = 42);
```

**3. No Database Enforcement of Data Consistency**

- Permission records can point to deleted entities
- IDs can be completely invalid (negative, zero, non-existent)
- The database has no mechanism to validate the type/ID combination
- Data integrity relies entirely on application logic

#### ✅ What You Gain

**1. Flexibility and Simplicity**

One unified table handles all four permission types:
- Simpler schema (4-in-1 instead of 4 separate tables)
- Unified querying (one table to check permissions)
- Easy to add new permission types in the future

**2. Performance Benefits**

- No foreign key constraint checking on INSERT/UPDATE operations
- Faster writes (though the difference is usually marginal)
- Simplified transaction handling

**3. Decoupled Lifecycle**

- Permissions can exist temporarily before entities are created
- Audit trails can be maintained even after entities are deleted
- More flexibility in data import/migration scenarios

### How This Design Compensates

The system mitigates the lack of foreign keys through **application-level validation** and **careful API design**:

#### 1. Validation Through the Repository Interface

The repository methods expect entities to exist before creating permissions. Looking at the API methods:

```go
// Stage 5 Interface - All methods accept entity IDs
func AddUserToUserPermission(ctx context.Context, sourceUserID, targetUserID int) error
func AddUserToUserGroupPermission(ctx context.Context, sourceUserID, targetUserGroupID int) error
func AddUserGroupToUserPermission(ctx context.Context, sourceUserGroupID, targetUserID int) error
func AddUserGroupToUserGroupPermission(ctx context.Context, sourceUserGroupID, targetUserGroupID int) error
```

#### 2. Creation Order in Tests and Integration

All tests follow the pattern: create entities first, then create permissions:

```go
// Entities must be created before permissions
alice, _ := srv.CreateUser(ctx, "Alice")
bob, _ := srv.CreateUser(ctx, "Bob")
admins, _ := srv.CreateUserGroup(ctx, "Admins")

// Now permissions can be created
srv.AddUserToUserPermission(ctx, alice, bob)          // User → User
srv.AddUserGroupToUserPermission(ctx, admins, bob)    // Group → User
```

#### 3. Graceful Handling of Missing References

Permission queries simply return no matches if referenced entities don't exist:

```sql
-- This query won't fail if IDs are invalid, it just returns empty
SELECT 1 FROM permissions p
INNER JOIN users u ON u.id = ?
WHERE p.source_type = 'user' AND p.source_id = ?
```

The system degrades gracefully - orphaned permissions are ignored rather than causing errors.

#### 4. Controlled Access Pattern

- The database is not exposed directly to users
- All operations go through the Go API
- The API can enforce existence checks before creating permissions
- No ad-hoc SQL manipulation of the permissions table

### Alternative Designs Considered

#### Option 1: Four Separate Permission Tables

Split permissions into four tables with proper foreign keys:

```sql
CREATE TABLE user_to_user_permissions (
    source_user_id INT NOT NULL,
    target_user_id INT NOT NULL,
    PRIMARY KEY (source_user_id, target_user_id),
    FOREIGN KEY (source_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (target_user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE user_to_group_permissions (
    source_user_id INT NOT NULL,
    target_group_id INT NOT NULL,
    PRIMARY KEY (source_user_id, target_group_id),
    FOREIGN KEY (source_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (target_group_id) REFERENCES user_groups(id) ON DELETE CASCADE
);

CREATE TABLE group_to_user_permissions (
    source_group_id INT NOT NULL,
    target_user_id INT NOT NULL,
    PRIMARY KEY (source_group_id, target_user_id),
    FOREIGN KEY (source_group_id) REFERENCES user_groups(id) ON DELETE CASCADE,
    FOREIGN KEY (target_user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE group_to_group_permissions (
    source_group_id INT NOT NULL,
    target_group_id INT NOT NULL,
    PRIMARY KEY (source_group_id, target_group_id),
    FOREIGN KEY (source_group_id) REFERENCES user_groups(id) ON DELETE CASCADE,
    FOREIGN KEY (target_group_id) REFERENCES user_groups(id) ON DELETE CASCADE
);
```

**Advantages:**
- ✅ Full referential integrity with foreign keys
- ✅ Automatic cascade deletes
- ✅ Type safety enforced at database level
- ✅ Cannot insert invalid references

**Disadvantages:**
- ❌ 4 tables instead of 1 (schema complexity)
- ❌ Permission queries must UNION across all 4 tables
- ❌ More complex application code (must route to correct table)
- ❌ Harder to add new entity types
- ❌ More maintenance burden
- ❌ Larger schema surface area

**Why Rejected:**
The complexity cost outweighs the referential integrity benefit, especially since the application layer already validates entity existence.

#### Option 2: CHECK Constraints with Subqueries

Use CHECK constraints to validate references:

```sql
CREATE TABLE permissions (
    source_type ENUM('user', 'group') NOT NULL,
    source_id INT NOT NULL,
    target_type ENUM('user', 'group') NOT NULL,
    target_id INT NOT NULL,
    CHECK (
        (source_type = 'user' AND EXISTS (SELECT 1 FROM users WHERE id = source_id))
        OR
        (source_type = 'group' AND EXISTS (SELECT 1 FROM user_groups WHERE id = source_id))
    ),
    -- Similar CHECK for target_type/target_id
    PRIMARY KEY (source_type, source_id, target_type, target_id)
);
```

**Why Rejected:**
- ❌ Not supported in MySQL (CHECK constraints can't contain subqueries)
- ❌ Significant performance overhead (runs on every INSERT/UPDATE)
- ❌ No automatic cascade behavior
- ❌ More complex to maintain than foreign keys

#### Option 3: Database Triggers

Use triggers to validate references and handle cascades:

```sql
DELIMITER $$
CREATE TRIGGER validate_permission_source
BEFORE INSERT ON permissions
FOR EACH ROW
BEGIN
    IF NEW.source_type = 'user' THEN
        IF NOT EXISTS (SELECT 1 FROM users WHERE id = NEW.source_id) THEN
            SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Invalid user reference';
        END IF;
    ELSEIF NEW.source_type = 'group' THEN
        IF NOT EXISTS (SELECT 1 FROM user_groups WHERE id = NEW.source_id) THEN
            SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'Invalid group reference';
        END IF;
    END IF;
END$$
DELIMITER ;
```

**Why Rejected:**
- ❌ Complex and error-prone to maintain
- ❌ Poor performance (triggers run on every operation)
- ❌ Hard to debug
- ❌ Would need multiple triggers (INSERT, UPDATE, DELETE on users/groups)
- ❌ Adds significant database complexity
- ❌ Not idiomatic for this type of validation

### Trade-offs Analysis

#### Advantages of No Foreign Keys (Current Choice)

✅ **Simple, unified schema** - One table, clear structure  
✅ **Easier permission queries** - Single table to check  
✅ **Flexible design** - Easy to extend with new entity types  
✅ **No query overhead** - Foreign key checks have cost  
✅ **Application-controlled lifecycle** - Explicit about what happens on delete  
✅ **Appropriate for the use case** - Controlled API access pattern  

#### Disadvantages Accepted

⚠️ **No database-enforced referential integrity** - Must rely on application  
⚠️ **Potential for orphaned records** - No automatic cleanup  
⚠️ **Manual cascade handling** - Application must clean up permissions  
⚠️ **Risk of invalid data** - If application validation is bypassed  

### Mitigation Strategies

**How the system handles the disadvantages:**

1. **Controlled Access Layer**
   - All database operations go through the repository interface
   - No direct SQL access in production
   - API validates entity existence before creating permissions

2. **Comprehensive Testing**
   - 24 unit tests verify correct behavior
   - Integration tests validate end-to-end scenarios
   - Tests ensure entities are created before permissions

3. **Clear Documentation**
   - API design makes the creation order obvious
   - Code examples in README show correct patterns
   - This document explains the design decisions

4. **Graceful Degradation**
   - Queries handle missing references without errors
   - Permission checks simply return "no permission" for invalid references
   - System remains stable even with orphaned data

### When to Reconsider This Decision

Consider adding foreign keys (via separate tables) if:

- 📦 **Multiple systems** directly write to the database
- 📦 **Manual SQL operations** become common
- 📦 **Data integrity** becomes critical (financial, medical, legal)
- 📦 **Audit requirements** demand database-level enforcement
- 📦 **Application validation** proves unreliable or inconsistent
- 📦 **Production incidents** occur due to orphaned permissions
- 📦 **The team** strongly prefers database-enforced constraints

Consider adding triggers if:
- 📦 You need referential integrity but can't change the table structure
- 📦 Performance impact is acceptable
- 📦 Team has expertise in trigger maintenance

### Is This Design a Problem?

**For this system: No, it's acceptable because:**

1. ✅ Controlled API prevents direct database manipulation
2. ✅ Application layer validates entity existence
3. ✅ Tests verify correct usage patterns
4. ✅ The system degrades gracefully with invalid references
5. ✅ It's a permissions system, not a financial ledger
6. ✅ The flexibility benefits outweigh the integrity risks

**This would be problematic if:**

- ❌ Direct SQL access to the database is common
- ❌ Multiple applications write to the same database
- ❌ No application-level validation exists
- ❌ Data consistency is legally or financially critical
- ❌ Debugging orphaned permissions becomes a frequent issue

### Implementation Notes

**Current Implementation:**
- Polymorphic type+id pattern in permissions table
- No foreign key constraints
- Application-level validation through the repository
- Graceful handling of missing references in queries

**Future Improvements (if needed):**
- Add application-level cleanup of orphaned permissions
- Implement soft deletes for users/groups to preserve permission history
- Add database views that validate permission references
- Create monitoring/alerts for orphaned permission records

### References

- [Polymorphic Associations](https://stackoverflow.com/questions/922184/why-can-you-not-have-a-foreign-key-in-a-polymorphic-association)
- [Database Foreign Keys vs Application Logic](https://stackoverflow.com/questions/18717/to-use-or-not-to-use-database-foreign-key-constraints)
- [Polymorphic Associations in Rails (relevant pattern)](https://guides.rubyonrails.org/association_basics.html#polymorphic-associations)

### Conclusion

The permissions table's lack of foreign keys is a **deliberate design choice** driven by the polymorphic nature of permission relationships. SQL databases cannot enforce conditional foreign keys that reference different tables based on a type discriminator. While this means losing database-level referential integrity, the system compensates through application-level validation, controlled API access, and graceful error handling.

The trade-off is appropriate for this use case: the flexibility and simplicity of a unified permissions table outweigh the benefits of database-enforced referential integrity, especially given the controlled access patterns and comprehensive testing. Alternative designs (four separate tables, triggers, CHECK constraints) would add significant complexity without meaningful benefits for this permission management system.

This design should be reconsidered only if direct database access becomes common, multiple systems need to write to the database, or production incidents reveal that orphaned permissions are causing operational problems.

