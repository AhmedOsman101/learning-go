# Phase 06: Advanced Concurrency Patterns

**Duration**: 2-3 weeks  
**Prerequisites**: Phase 05 completed  
**Practice Directory**: `phase-06-concurrency-patterns/`

## Overview

Building on the fundamentals, this phase explores advanced concurrency patterns used in production Go applications. You'll learn about context for cancellation, sophisticated channel patterns, and how to build robust concurrent systems.

## Learning Objectives

- Master the context package for cancellation and timeouts
- Implement advanced channel patterns
- Build worker pools with graceful shutdown
- Handle backpressure and rate limiting
- Debug concurrent programs
- Write concurrent code that's easy to reason about

## Topics to Cover

### 1. Context Package

```go
import "context"

// Context carries deadlines, cancellation, and request-scoped values
// across API boundaries and between goroutines.

// Creating contexts
ctx := context.Background()  // Root context, never cancelled
ctx := context.TODO()        // When unsure which context to use

// With cancellation
ctx, cancel := context.WithCancel(parentCtx)
cancel()  // Call when done to release resources

// With timeout
ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
defer cancel()

// With deadline
deadline := time.Now().Add(5 * time.Second)
ctx, cancel := context.WithDeadline(parentCtx, deadline)
defer cancel()

// With value
ctx := context.WithValue(parentCtx, key, value)
value := ctx.Value(key)

// Using context in functions
func DoWork(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()  // context.Canceled or context.DeadlineExceeded
        default:
            // Do work
        }
    }
}

// HTTP request with context
func FetchURL(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    return io.ReadAll(resp.Body)
}

// Database query with context
func QueryUser(ctx context.Context, db *sql.DB, id int) (*User, error) {
    row := db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?", id)
    // ...
}

// Propagating context through layers
func Handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Add request ID
    requestID := generateID()
    ctx = context.WithValue(ctx, "requestID", requestID)

    // Add timeout
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()

    result, err := Service(ctx, ...)
    // ...
}

func Service(ctx context.Context, ...) (Result, error) {
    // Context is propagated automatically
    return Repository(ctx, ...)
}

func Repository(ctx context.Context, ...) (Result, error) {
    // Use ctx for cancellation
    return db.QueryContext(ctx, ...)
}
```

**Context Rules**:

1. Don't store contexts in structs - pass as first parameter
2. Don't pass nil context - use `context.TODO()` if unsure
3. Context.Value is for request-scoped data, not optional parameters
4. Always call cancel function to release resources

### 2. Graceful Shutdown

```go
func main() {
    // Create listener
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatal(err)
    }

    // Create server
    server := &http.Server{Handler: handler}

    // Channel to listen for shutdown signals
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    // Start server in goroutine
    errCh := make(chan error, 1)
    go func() {
        errCh <- server.Serve(listener)
    }()

    // Wait for signal or error
    select {
    case err := <-errCh:
        log.Printf("server error: %v", err)
    case sig := <-quit:
        log.Printf("received signal: %v", sig)
    }

    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Printf("shutdown error: %v", err)
        server.Close()  // Force close
    }
}

// Worker pool with graceful shutdown
type WorkerPool struct {
    workers   int
    taskCh    chan Task
    quit      chan struct{}
    wg        sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
    return &WorkerPool{
        workers: workers,
        taskCh:  make(chan Task, 100),
        quit:    make(chan struct{}),
    }
}

func (p *WorkerPool) Start() {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker()
    }
}

func (p *WorkerPool) worker() {
    defer p.wg.Done()
    for {
        select {
        case task := <-p.taskCh:
            task.Execute()
        case <-p.quit:
            // Drain remaining tasks
            for task := range p.taskCh {
                task.Execute()
            }
            return
        }
    }
}

func (p *WorkerPool) Stop() {
    close(p.quit)
    p.wg.Wait()
}

func (p *WorkerPool) Submit(task Task) {
    select {
    case p.taskCh <- task:
    case <-p.quit:
        // Pool is shutting down
    }
}
```

### 3. Advanced Channel Patterns

```go
// 1. Or-done channel
// Returns a channel that closes when any input channel closes
func OrDone(done <-chan struct{}, c <-chan interface{}) <-chan interface{} {
    valStream := make(chan interface{})
    go func() {
        defer close(valStream)
        for {
            select {
            case <-done:
                return
            case v, ok := <-c:
                if !ok {
                    return
                }
                select {
                case valStream <- v:
                case <-done:
                }
            }
        }
    }()
    return valStream
}

// 2. Tee channel (split one channel into two)
func Tee(done <-chan struct{}, in <-chan interface{}) (_, _ <-chan interface{}) {
    out1 := make(chan interface{})
    out2 := make(chan interface{})

    go func() {
        defer close(out1)
        defer close(out2)

        for val := range OrDone(done, in) {
            var out1, out2 = out1, out2
            for i := 0; i < 2; i++ {
                select {
                case <-done:
                    return
                case out1 <- val:
                    out1 = nil
                case out2 <- val:
                    out2 = nil
                }
            }
        }
    }()

    return out1, out2
}

// 3. Bridge channel (flatten channel of channels)
func Bridge(done <-chan struct{}, chanStream <-chan <-chan interface{}) <-chan interface{} {
    valStream := make(chan interface{})
    go func() {
        defer close(valStream)
        for {
            var stream <-chan interface{}
            select {
            case maybeStream, ok := <-chanStream:
                if !ok {
                    return
                }
                stream = maybeStream
            case <-done:
                return
            }

            for val := range OrDone(done, stream) {
                select {
                case valStream <- val:
                case <-done:
                }
            }
        }
    }()
    return valStream
}

// 4. Queue channel (unbounded buffer)
func Queue(done <-chan struct{}) (<-chan interface{}, func(interface{})) {
    in := make(chan interface{})
    out := make(chan interface{})

    go func() {
        var queue []interface{}
        var next interface{}
        var outCh chan interface{}

        defer close(out)

        for {
            if len(queue) == 0 {
                outCh = nil
            } else {
                outCh = out
                next = queue[0]
            }

            select {
            case <-done:
                return
            case v, ok := <-in:
                if !ok {
                    return
                }
                queue = append(queue, v)
            case outCh <- next:
                queue = queue[1:]
            }
        }
    }()

    return out, func(v interface{}) {
        in <- v
    }
}

// 5. Heartbeat pattern
func DoWorkWithHeartbeat(done <-chan struct{}, pulseInterval time.Duration) (<-chan interface{}, <-chan time.Time) {
    heartbeat := make(chan interface{})
    results := make(chan time.Time)

    go func() {
        defer close(heartbeat)
        defer close(results)

        pulse := time.NewTicker(pulseInterval)
        workDone := make(chan time.Time)

        go func() {
            for {
                select {
                case <-done:
                    return
                case workDone <- time.Now():
                }
            }
        }()

        for {
            select {
            case <-done:
                return
            case <-pulse.C:
                select {
                case heartbeat <- struct{}{}:
                default:
                }
            case r := <-workDone:
                select {
                case results <- r:
                case <-done:
                    return
                }
            }
        }
    }()

    return heartbeat, results
}
```

### 4. Rate Limiting Patterns

```go
// 1. Simple rate limiter
type RateLimiter struct {
    ticker *time.Ticker
    tokens chan struct{}
}

func NewRateLimiter(rate int, burst int) *RateLimiter {
    rl := &RateLimiter{
        ticker: time.NewTicker(time.Second / time.Duration(rate)),
        tokens: make(chan struct{}, burst),
    }

    // Fill initial tokens
    for i := 0; i < burst; i++ {
        rl.tokens <- struct{}{}
    }

    go func() {
        for range rl.ticker.C {
            select {
            case rl.tokens <- struct{}{}:
            default:
            }
        }
    }()

    return rl
}

func (rl *RateLimiter) Wait() {
    <-rl.tokens
}

func (rl *RateLimiter) WaitContext(ctx context.Context) error {
    select {
    case <-rl.tokens:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

// 2. Token bucket with golang.org/x/time/rate
import "golang.org/x/time/rate"

func rateLimitedClient() {
    limiter := rate.NewLimiter(rate.Limit(100), 10)  // 100/sec, burst 10

    for _, url := range urls {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

        err := limiter.Wait(ctx)
        if err != nil {
            cancel()
            continue
        }

        go fetchURL(ctx, url)
        cancel()
    }
}

// 3. Per-client rate limiting
type IPRateLimiter struct {
    mu       sync.Mutex
    limiters map[string]*rate.Limiter
    rate     rate.Limit
    burst    int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
    return &IPRateLimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     r,
        burst:    b,
    }
}

func (rl *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    limiter, exists := rl.limiters[ip]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.limiters[ip] = limiter
    }

    return limiter
}

// Middleware
func RateLimitMiddleware(rl *IPRateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := r.RemoteAddr
            limiter := rl.GetLimiter(ip)

            if !limiter.Allow() {
                http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### 5. Circuit Breaker Pattern

```go
type CircuitBreaker struct {
    maxFailures   int
    timeout       time.Duration
    state         State
    failures      int
    lastFailTime  time.Time
    mu            sync.Mutex
}

type State int

const (
    StateClosed State = iota
    StateOpen
    StateHalfOpen
)

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        maxFailures: maxFailures,
        timeout:     timeout,
        state:       StateClosed,
    }
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    cb.mu.Lock()

    switch cb.state {
    case StateOpen:
        if time.Since(cb.lastFailTime) > cb.timeout {
            cb.state = StateHalfOpen
        } else {
            cb.mu.Unlock()
            return errors.New("circuit breaker is open")
        }
    }

    cb.mu.Unlock()

    err := fn()

    cb.mu.Lock()
    defer cb.mu.Unlock()

    if err != nil {
        cb.failures++
        cb.lastFailTime = time.Now()

        if cb.failures >= cb.maxFailures {
            cb.state = StateOpen
        }
        return err
    }

    cb.failures = 0
    cb.state = StateClosed
    return nil
}

// Usage
var breaker = NewCircuitBreaker(5, 30*time.Second)

func callExternalService() error {
    return breaker.Call(func() error {
        resp, err := http.Get("http://external-service/api")
        if err != nil {
            return err
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 500 {
            return errors.New("server error")
        }
        return nil
    })
}
```

### 6. Error Group

```go
import "golang.org/x/sync/errgroup"

// Basic usage
func fetchAll(urls []string) ([]string, error) {
    g, ctx := errgroup.WithContext(context.Background())
    results := make([]string, len(urls))

    for i, url := range urls {
        i, url := i, url
        g.Go(func() error {
            // Context is cancelled if any goroutine fails
            data, err := fetchURL(ctx, url)
            if err != nil {
                return err
            }
            results[i] = data
            return nil
        })
    }

    if err := g.Wait(); err != nil {
        return nil, err
    }

    return results, nil
}

// With parallelism limit
func fetchAllLimited(urls []string, maxParallel int) ([]string, error) {
    g, ctx := errgroup.WithContext(context.Background())
    g.SetLimit(maxParallel)

    results := make([]string, len(urls))

    for i, url := range urls {
        i, url := i, url
        g.Go(func() error {
            data, err := fetchURL(ctx, url)
            if err != nil {
                return err
            }
            results[i] = data
            return nil
        })
    }

    return results, g.Wait()
}
```

### 7. Debugging Concurrency

```go
// 1. Deadlock detection
// Go runtime detects some deadlocks automatically
// "fatal error: all goroutines are asleep - deadlock!"

// 2. Race detection
// go run -race main.go
// go test -race ./...

// 3. Goroutine leak detection
import "runtime"

func printGoroutines() {
    fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine())
}

// 4. Stack traces
func dumpGoroutines() {
    buf := make([]byte, 1<<20)
    n := runtime.Stack(buf, true)
    fmt.Printf("%s\n", buf[:n])
}

// 5. HTTP pprof
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    // ...
}

// View goroutines: http://localhost:6060/debug/pprof/goroutine
// Command line: go tool pprof http://localhost:6060/debug/pprof/goroutine

// 6. Trace
import "runtime/trace"

func main() {
    f, err := os.Create("trace.out")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    trace.Start(f)
    defer trace.Stop()

    // Your program...
}

// View trace: go tool trace trace.out
```

## Hands-on Exercises

Create the following programs in `phase-06-concurrency-patterns/`:

### Exercise 1: Context-Aware HTTP Client

Build an HTTP client that:

- Respects context cancellation
- Supports retries with backoff
- Has configurable timeouts
- Logs request IDs from context

### Exercise 2: Graceful Server

Create an HTTP server that:

- Handles shutdown signals
- Waits for in-flight requests
- Has configurable shutdown timeout
- Logs shutdown progress

### Exercise 3: Pipeline System

Build a data processing pipeline:

- Multiple stages (fetch, parse, transform, store)
- Each stage can fail independently
- Supports backpressure
- Has metrics at each stage

### Exercise 4: Rate-Limited API Gateway

Create an API gateway that:

- Rate limits per client
- Has circuit breaker for backends
- Supports graceful degradation
- Logs all events

## Resources

### Official

- [Context Package](https://pkg.go.dev/context)
- [Go Concurrency Patterns: Context](https://go.dev/blog/context)
- [pprof](https://pkg.go.dev/runtime/pprof)

### Deep Dives

- [Concurrency in Go](https://www.oreilly.com/library/view/concurrency-in-go/9781491941294/) (Book)
- [Advanced Patterns Blog](https://blog.golang.org/pipelines)

## Validation Checklist

- [ ] Can use context for cancellation
- [ ] Understand context propagation
- [ ] Can implement graceful shutdown
- [ ] Know advanced channel patterns
- [ ] Can implement rate limiting
- [ ] Understand circuit breaker pattern
- [ ] Can debug concurrent programs
- [ ] All exercises completed

## Common Pitfalls

| Mistake                   | Consequence       | Fix                     |
| ------------------------- | ----------------- | ----------------------- |
| Not calling cancel        | Resource leak     | Always defer cancel()   |
| Storing context in struct | Unclear ownership | Pass as parameter       |
| Ignoring ctx.Done()       | Goroutine leak    | Check in loops          |
| Not propagating context   | Lost cancellation | Pass through all layers |

## Next Phase

Proceed to **Phase 07: Standard Library Deep Dive** to master Go's powerful standard library.
