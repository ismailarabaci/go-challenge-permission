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

