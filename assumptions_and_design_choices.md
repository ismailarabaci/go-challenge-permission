# Assumptions and Design Choices

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

- 32 total test functions organized by stage and scenario
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

All 32 tests passing (24 unit tests + 2 integration tests with 11 subtests)
- Total execution time: ~2 seconds
- All scenarios covered: user operations, groups, hierarchies, transitive membership, permissions
- Integration tests validate full HTTP request/response cycle

