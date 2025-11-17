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
✅ CI pipeline with GitHub Actions for automated testing and quality checks

### Conclusion

AI-assisted development with **critical code review** enabled rapid learning and high-quality implementation. Success came from active engagement—scrutinizing, questioning, and adjusting AI-generated code rather than blind acceptance. AI served as an intelligent pair programmer that could be questioned, guided, and validated through active engagement.

---

## Synchronous vs Asynchronous Database Operations

### Decision
The repository layer uses **synchronous database calls** rather than asynchronous patterns.

### Rationale

Go's concurrency model differs fundamentally from JavaScript or Python:

1. **Built-in Concurrency:** Go's `net/http` spawns a goroutine per request automatically—all requests are already concurrent
2. **Connection Pooling:** The `database/sql` package provides automatic, thread-safe connection pooling
3. **Idiomatic Go:** Synchronous code is the Go community standard for repositories—easier to read, test, and debug

### Trade-offs

**Advantages (current choice):** Simple readable code, easy error handling, idiomatic, already concurrent at request level, context-based cancellation

**Alternatives rejected:** Async patterns add complexity (channels, synchronization) without performance benefit in Go's model

### Conclusion
The synchronous pattern leverages Go's goroutines-per-request and connection pooling for excellent performance without async complexity.

## Package Structure: Single `server` Package vs Separate `repository` Package

### Decision
All code in a single **`pkg/server` package** organized into separate files rather than separate packages.

**Structure:** `interface.go`, `server.go`, `repository.go`, `mysql_repository.go`, `errors.go`, `config.go`, `server_test.go`

### Rationale

**Assignment Requirements:** Challenge requires implementing interfaces in `pkg/server/interface.go`

**Go Best Practices:**
- Start simple, refactor when needed—avoid premature abstraction
- Package by feature, not by layer (modern Go convention)
- Standard library precedent (`net/http`, `database/sql` are large cohesive packages)

**Benefits:** Low ceremony (no exported types for internal use), easy navigation, clear file separation, appropriate scope (~500-700 lines), Repository interface enables testing

**When to reconsider:** Multiple packages need the repository, codebase grows to thousands of lines, or publishing as separate module

### Conclusion
Single-package structure is appropriate for this focused coding challenge, balancing simplicity with good engineering practices while following Go idioms.

---

## Constructor Pattern: Enforcing Pure Dependency Injection

### Decision
Single `New(repo Repository) *Server` constructor requiring **explicit dependency injection**—no convenience constructors.

**Factory functions provided:**
- `OpenDatabase(config Config) (*sql.DB, error)`
- `NewMySQLRepository(db *sql.DB) *MySQLRepository`
- `DefaultConfig() Config`

### Rationale

**Explicit over implicit:** Dependencies always visible, no hidden resource allocation, clear dependency graph

**Testability first:** Mock injection straightforward, test setup explicit about what it creates

**Production-ready:** Supports custom configuration, different environments, different repository implementations

**Usage pattern:**
```go
config := server.DefaultConfig()
db, _ := server.OpenDatabase(config)
repo := server.NewMySQLRepository(db)
srv := server.New(repo)
defer srv.Close()
```

### Trade-offs

**Advantages:** Explicit dependencies, maximum testability, flexibility, no hidden side effects, simplest API

**Cost accepted:** More verbose (4 lines vs 1), requires understanding dependency chain—but this verbosity is self-documenting

**Alternatives rejected:** Convenience constructors hide dependencies and reduce flexibility; multiple constructors create confusion; builder pattern over-engineers

### Conclusion
Pure dependency injection prioritizes explicitness, testability, and flexibility over convenience. Pattern aligns with Go idioms and makes the dependency graph transparent.

---

## Test Structure and Organization

### Decision
Granular unit tests with table-driven patterns plus integration tests for end-to-end validation.

**Test Suite:** 24 unit tests (`server_test.go`) + 2 integration tests (`integration_test.go` with 11 subtests)

### Approach

**Unit Tests:** Organized by stage (`Test_Stage{X}_{Description}`), focused on specific scenarios, table-driven where applicable

**Integration Tests:** End-to-end HTTP testing with handlers in test file, covers complex permission scenarios with nested groups

**Go Patterns Used:** Table-driven tests, subtests with `t.Run()`, test helpers with `t.Helper()`, `httptest.Server`, proper cleanup with `defer`

### Rationale

**Benefits:** Each test covers specific scenario (easy failure identification), table-driven tests simplify adding cases, descriptive names are self-documenting, helper functions eliminate duplication

**Why HTTP in tests:** Validates realistic usage through HTTP layer without requiring production HTTP code, demonstrates real-world application patterns

### Test Results
All tests passing in ~2 seconds, covering user operations, groups, hierarchies, transitive membership, and permissions with both positive and negative cases

---

## Continuous Integration (CI) Pipeline

### Decision
GitHub Actions for automated testing, linting, and code quality checks.

### Configuration

**Three jobs** run on every push/PR:
1. **Test:** Multi-version testing (Go 1.21-1.23) with MySQL 8.0 service container, race detector, coverage reporting to Codecov
2. **Lint:** golangci-lint with comprehensive linters (errcheck, gosec, govet, staticcheck, gofmt, ineffassign, dupl)
3. **Format:** Code style enforcement with `gofmt` and `go vet`

**Features:** Matrix builds for version compatibility, service containers mirror production, caching speeds builds (~2-3 min runs), Dependabot for automated updates

### Rationale

**Why GitHub Actions:** Native integration, free for public repos, easy service containers, parallel matrix builds

**Benefits:** Catches issues before merge, prevents regressions, documents build process, provides visibility via badges

---

## Database Schema: Separate Tables vs Single Principals Table

### Decision
Separate **`users` and `user_groups` tables** rather than a unified `principals` table with type discriminator.

### Rationale

**Type Safety:** Separate tables enable foreign key constraints that enforce type safety—`user_group_members` can ONLY reference actual users and groups, preventing invalid relationships. A unified `principals` table cannot enforce this at the database level.

**Simpler Queries:** No type discriminators needed in queries. `SELECT name FROM users WHERE id = ?` vs `SELECT name FROM principals WHERE id = ? AND type = 'user'`

**Semantic Clarity:** Users and groups are fundamentally different—users can be members of groups, groups can contain users and other groups. The API (Stage interfaces) treats them differently, schema should match.

**Database-Level Constraints:** Foreign keys and CHECK constraints enforce business rules impossible to violate: only users in `user_group_members`, only groups in `user_group_hierarchy`

**Performance:** Smaller tables, more efficient indexes, better query optimization without type filters

### Trade-offs

**Advantages:** Database-enforced type safety, simpler queries, better performance, clear semantics, standard well-understood pattern

**Cost accepted:** Two tables instead of one (minimal overhead), separate ID spaces (actually a feature for type safety)

**Alternatives rejected:** Single principals table (loses type safety, complex queries), class table inheritance (over-engineering), separate permission tables (schema explosion)

### Conclusion
Separate tables leverage database constraints for business rule enforcement, simpler queries, and accurate domain modeling. The permissions table correctly uses polymorphism where needed (any source → any target), but this doesn't mean principals themselves should be unified.

---

## Permissions Table: No Foreign Key Constraints

### Decision
Polymorphic pattern with type discriminators (`source_type`, `target_type`) but **no foreign key constraints**.

**The polymorphic pattern:** Same ID column references different tables based on type—`source_id` can reference `users.id` or `user_groups.id` depending on `source_type`. Creates four permission types: User→User, User→Group, Group→User, Group→Group.

### Why No Foreign Keys

SQL foreign keys **cannot be conditional**—they must reference a single, static table. Cannot write `FOREIGN KEY (source_id) REFERENCES CASE WHEN source_type='user' THEN users(id)...` (this syntax doesn't exist).

### Trade-offs

**What you lose:** No database-level referential integrity, no automatic cascade deletes, can insert invalid references, orphaned permissions possible

**What you gain:** Simple unified schema (one table handles all 4 types), unified querying, flexible design, no FK overhead

**How it compensates:** Application-level validation through repository interface, controlled API access (no direct SQL), tests ensure entities created before permissions, graceful degradation (invalid references ignored in queries)

**Alternatives rejected:** Four separate tables with FKs (schema complexity, UNION queries, maintenance burden), CHECK constraints (not supported in MySQL), triggers (complex, poor performance)

### When to Reconsider

Consider foreign keys (via separate tables) if: multiple systems write to DB, manual SQL operations become common, data integrity becomes critical (financial/medical/legal), or production incidents occur from orphaned permissions

### Conclusion
Deliberate choice driven by polymorphic requirements. SQL cannot enforce conditional FKs. Application-level validation, controlled access, and comprehensive testing compensate for lack of database-enforced referential integrity. Flexibility and simplicity outweigh integrity trade-offs for this use case.

