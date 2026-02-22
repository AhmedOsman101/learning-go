# Phase 05: Concurrency Fundamentals

**Duration**: 2-3 weeks  
**Prerequisites**: Phase 04 completed  
**Practice Directory**: `phase-05-concurrency-basics/`

## Overview

Concurrency is Go's killer feature. Unlike other languages that bolt on concurrency, Go was designed with concurrency in mind. This phase covers goroutines, channels, the select statement, and basic synchronization primitives.

## Learning Objectives

- Understand goroutines and the Go scheduler
- Master channel operations (send, receive, close)
- Use buffered vs unbuffered channels appropriately
- Apply the select statement for channel multiplexing
- Use sync primitives (Mutex, WaitGroup, Once)
- Understand common concurrency patterns

## Topics to Cover

### 1. Goroutines

```go
// Starting a goroutine
go func() {
    fmt.Println("Running in goroutine")
}()

// Goroutine with parameters
go func(msg string) {
    fmt.Println(msg)
}("hello")

// Wait for goroutine to finish
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    doWork()
}()
wg.Wait()

// Goroutine leak prevention
func worker(done <-chan struct{}) {
    for {
        select {
        case <-done:
            return
        default:
            // Do work
        }
    }
}

// Capturing loop variables (common bug!)
for _, item := range items {
    go func() {
        process(item)  // BUG: all goroutines use last item
    }()
}

// Correct approach
for _, item := range items {
    go func(i Item) {
        process(i)
    }(item)
}

// Or in Go 1.22+
for _, item := range items {
    go func() {
        process(item)  // OK in Go 1.22+
    }()
}
```

**Goroutine Internals**:

- M:N scheduler (M goroutines on N OS threads)
- Green threads - lightweight (~2KB stack initially)
- Grows/shrinks as needed
- Cooperative scheduling (preemption in newer Go)

### 2. Channels

```go
// Unbuffered channel
ch := make(chan int)

// Buffered channel
ch := make(chan int, 10)

// Send
ch <- 42

// Receive
value := <-ch
value, ok := <-ch  // ok is false if channel closed

// Close channel (sender only!)
close(ch)

// Range over channel
for value := range ch {
    fmt.Println(value)
}

// Channel directions in function signatures
func sender(ch chan<- int) {    // Send only
    ch <- 42
}

func receiver(ch <-chan int) {  // Receive only
    value := <-ch
}

func bidirectional(ch chan int) {  // Both
    ch <- 42
    value := <-ch
}

// Nil channel behavior
var nilCh chan int
// Send blocks forever
// Receive blocks forever
// Close panics
// Useful for disabling select cases
```

**Channel Rules**:
| Operation | Unbuffered | Buffered | Closed | Nil |
|-----------|------------|----------|--------|-----|
| Send | Blocks until receive | Blocks if full | Panic | Blocks forever |
| Receive | Blocks until send | Blocks if empty | Zero value, false | Blocks forever |
| Close | Works | Works | Panic | Panic |

### 3. Select Statement

```go
// Basic select
select {
case msg := <-ch1:
    fmt.Println("from ch1:", msg)
case msg := <-ch2:
    fmt.Println("from ch2:", msg)
case ch3 <- 42:
    fmt.Println("sent to ch3")
}

// With default (non-blocking)
select {
case msg := <-ch:
    fmt.Println(msg)
default:
    fmt.Println("no message")
}

// Timeout pattern
select {
case msg := <-ch:
    fmt.Println(msg)
case <-time.After(time.Second):
    fmt.Println("timeout")
}

// Cancellation pattern
func worker(ctx context.Context, ch <-chan int) {
    for {
        select {
        case <-ctx.Done():
            return
        case value := <-ch:
            process(value)
        }
    }
}

// Random selection
// If multiple cases ready, Go picks randomly
select {
case ch1 <- 1:
case ch2 <- 2:
}

// Disable case with nil channel
func process(enabled bool, ch chan int) {
    var disabled chan int  // nil
    if enabled {
        disabled = ch
    }

    select {
    case disabled <- 42:  // Only works if enabled
        fmt.Println("sent")
    default:
        fmt.Println("disabled or full")
    }
}
```

### 4. Sync Package

```go
import "sync"

// Mutex - mutual exclusion
var mu sync.Mutex
var counter int

func increment() {
    mu.Lock()
    defer mu.Unlock()
    counter++
}

// RWMutex - read/write lock
var rwmu sync.RWMutex
var data map[string]string

func read(key string) string {
    rwmu.RLock()
    defer rwmu.RUnlock()
    return data[key]
}

func write(key, value string) {
    rwmu.Lock()
    defer rwmu.Unlock()
    data[key] = value
}

// WaitGroup - wait for multiple goroutines
var wg sync.WaitGroup

func processItems(items []Item) {
    for _, item := range items {
        wg.Add(1)
        go func(i Item) {
            defer wg.Done()
            process(i)
        }(item)
    }
    wg.Wait()
}

// Once - execute exactly once
var once sync.Once
var instance *Singleton

func GetInstance() *Singleton {
    once.Do(func() {
        instance = &Singleton{}
    })
    return instance
}

// Pool - object pool
var pool = sync.Pool{
    New: func() interface{} {
        return new(Buffer)
    },
}

func getBuffer() *Buffer {
    return pool.Get().(*Buffer)
}

func putBuffer(b *Buffer) {
    b.Reset()
    pool.Put(b)
}

// Cond - condition variable
var cond = sync.NewCond(&sync.Mutex{})
var ready bool

func waitForCondition() {
    cond.L.Lock()
    for !ready {
        cond.Wait()
    }
    cond.L.Unlock()
}

func signalCondition() {
    cond.L.Lock()
    ready = true
    cond.Broadcast()
    cond.L.Unlock()
}

// Map - concurrent map
var m sync.Map

func store(key, value string) {
    m.Store(key, value)
}

func load(key string) (string, bool) {
    v, ok := m.Load(key)
    if !ok {
        return "", false
    }
    return v.(string), true
}

// Range over sync.Map
m.Range(func(key, value interface{}) bool {
    fmt.Printf("%v: %v\n", key, value)
    return true  // continue iteration
})

// Atomic operations
import "sync/atomic"

var counter int64

func incrementAtomic() {
    atomic.AddInt64(&counter, 1)
}

func getValue() int64 {
    return atomic.LoadInt64(&counter)
}

func compareAndSwap(old, new int64) bool {
    return atomic.CompareAndSwapInt64(&counter, old, new)
}
```

### 5. Common Patterns

```go
// 1. Worker Pool
func workerPool(jobs <-chan Job, results chan<- Result, workers int) {
    var wg sync.WaitGroup

    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                results <- process(job)
            }
        }()
    }

    go func() {
        wg.Wait()
        close(results)
    }()
}

// 2. Fan-out, Fan-in
func fanOut(input <-chan int, n int) []<-chan int {
    channels := make([]<-chan int, n)
    for i := 0; i < n; i++ {
        channels[i] = worker(input)
    }
    return channels
}

func fanIn(channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for v := range c {
                out <- v
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}

// 3. Pipeline
func pipeline() {
    // Stage 1: Generate
    gen := func(nums ...int) <-chan int {
        out := make(chan int)
        go func() {
            defer close(out)
            for _, n := range nums {
                out <- n
            }
        }()
        return out
    }

    // Stage 2: Square
    square := func(in <-chan int) <-chan int {
        out := make(chan int)
        go func() {
            defer close(out)
            for n := range in {
                out <- n * n
            }
        }()
        return out
    }

    // Connect stages
    nums := gen(1, 2, 3, 4, 5)
    squares := square(nums)

    for s := range squares {
        fmt.Println(s)
    }
}

// 4. Cancellation
func operation(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Do work
        }
    }
}

// 5. Rate Limiting
func rateLimiter(requests <-chan Request, perSecond int) <-chan Response {
    out := make(chan Response)
    ticker := time.NewTicker(time.Second / time.Duration(perSecond))

    go func() {
        defer close(out)
        defer ticker.Stop()

        for req := range requests {
            <-ticker.C  // Wait for tick
            out <- process(req)
        }
    }()

    return out
}

// 6. Timeout
func withTimeout(timeout time.Duration) (Result, error) {
    resultCh := make(chan Result, 1)
    errCh := make(chan error, 1)

    go func() {
        result, err := slowOperation()
        if err != nil {
            errCh <- err
            return
        }
        resultCh <- result
    }()

    select {
    case result := <-resultCh:
        return result, nil
    case err := <-errCh:
        return Result{}, err
    case <-time.After(timeout):
        return Result{}, errors.New("timeout")
    }
}
```

### 6. Race Detection

```go
// Race condition example (BAD)
var counter int

func main() {
    for i := 0; i < 1000; i++ {
        go func() {
            counter++  // Race condition!
        }()
    }
    time.Sleep(time.Second)
    fmt.Println(counter)  // Unpredictable result
}

// Fix with mutex
var mu sync.Mutex
var counter int

func main() {
    for i := 0; i < 1000; i++ {
        go func() {
            mu.Lock()
            counter++
            mu.Unlock()
        }()
    }
    time.Sleep(time.Second)
    fmt.Println(counter)  // Always 1000
}

// Detect races at runtime
// go run -race main.go
// go test -race ./...
```

## Hands-on Exercises

Create the following programs in `phase-05-concurrency-basics/`:

### Exercise 1: Concurrent Web Scraper

Build a web scraper that:

- Fetches multiple URLs concurrently
- Limits concurrent requests
- Handles timeouts
- Collects results

### Exercise 2: Job Queue System

Implement a job queue with:

- Multiple workers
- Job prioritization
- Graceful shutdown
- Job retry logic

### Exercise 3: Pub/Sub System

Create a publish-subscribe system:

- Multiple topics
- Multiple subscribers per topic
- Message buffering
- Subscriber removal

### Exercise 4: Concurrent Cache

Build a thread-safe cache with:

- Read/write separation
- Expiration
- Background cleanup
- Metrics collection

## Resources

### Official

- [Effective Go: Concurrency](https://go.dev/doc/effective_go#concurrency)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Share Memory By Communicating](https://go.dev/blog/codelab-share)

### Deep Dives

- [Advanced Patterns](https://blog.golang.org/advanced-go-concurrency-patterns)
- [Go Scheduler](https://www.ardanlabs.com/blog/2018/08/scheduling-in-go-part1.html)

## Validation Checklist

- [ ] Can start goroutines correctly
- [ ] Understand channel blocking behavior
- [ ] Can use select for multiplexing
- [ ] Know when to use buffered vs unbuffered channels
- [ ] Can use sync primitives correctly
- [ ] Understand race conditions
- [ ] Can use -race detector
- [ ] All exercises completed

## Common Pitfalls

| Mistake                       | Consequence        | Fix                         |
| ----------------------------- | ------------------ | --------------------------- |
| Closing channel from receiver | Panic              | Only sender closes          |
| Goroutine leak                | Memory leak        | Use context or done channel |
| Capturing loop variable       | All use last value | Pass as argument            |
| Shared memory without lock    | Race condition     | Use mutex or channels       |

## Next Phase

Proceed to **Phase 06: Advanced Concurrency Patterns** to learn context, cancellation, and complex patterns.
