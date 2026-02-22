# Phase 12: Advanced Topics and Best Practices

**Duration**: 2-3 weeks  
**Prerequisites**: Phase 11 completed  
**Practice Directory**: `phase-12-advanced/`

## Overview

This final phase covers advanced topics that separate intermediate from expert Go developers. You'll learn about reflection, unsafe operations, cgo, performance optimization, and best practices for production systems.

## Learning Objectives

- Understand and use reflection appropriately
- Know when and how to use unsafe
- Interface with C code using cgo
- Optimize Go programs for performance
- Apply production best practices
- Understand Go runtime internals

## Topics to Cover

### 1. Reflection

```go
import "reflect"

// Type and Value
func inspect(x interface{}) {
    t := reflect.TypeOf(x)
    v := reflect.ValueOf(x)

    fmt.Printf("Type: %v\n", t)
    fmt.Printf("Kind: %v\n", t.Kind())
    fmt.Printf("Value: %v\n", v)
}

// Kind enumeration
/*
Bool, Int, Int8, Int16, Int32, Int64,
Uint, Uint8, Uint16, Uint32, Uint64, Uintptr,
Float32, Float64, Complex64, Complex128,
Array, Chan, Func, Interface, Map, Pointer, Slice, String, Struct,
UnsafePointer
*/

// Working with structs
type Person struct {
    Name string `json:"name" validate:"required"`
    Age  int    `json:"age" validate:"min=0"`
}

func inspectStruct(x interface{}) {
    t := reflect.TypeOf(x)
    v := reflect.ValueOf(x)

    if t.Kind() == reflect.Ptr {
        t = t.Elem()
        v = v.Elem()
    }

    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        value := v.Field(i)

        fmt.Printf("Field: %s, Type: %v, Value: %v, Tag: %v\n",
            field.Name, field.Type, value.Interface(), field.Tag)

        // Get specific tag
        jsonTag := field.Tag.Get("json")
    }
}

// Creating values with reflection
func createSlice(t reflect.Type, length, capacity int) reflect.Value {
    return reflect.MakeSlice(reflect.SliceOf(t), length, capacity)
}

func createMap(keyType, valueType reflect.Type) reflect.Value {
    return reflect.MakeMap(reflect.MapOf(keyType, valueType))
}

// Calling functions
func callFunction(fn interface{}, args ...interface{}) []interface{} {
    v := reflect.ValueOf(fn)
    in := make([]reflect.Value, len(args))
    for i, arg := range args {
        in[i] = reflect.ValueOf(arg)
    }

    out := v.Call(in)
    result := make([]interface{}, len(out))
    for i, v := range out {
        result[i] = v.Interface()
    }
    return result
}

// Deep equality
reflect.DeepEqual(a, b)

// Practical example: Generic map function
func Map(slice, fn interface{}) interface{} {
    sliceVal := reflect.ValueOf(slice)
    fnVal := reflect.ValueOf(fn)

    if sliceVal.Kind() != reflect.Slice {
        panic("first argument must be a slice")
    }
    if fnVal.Kind() != reflect.Func {
        panic("second argument must be a function")
    }

    resultSlice := reflect.MakeSlice(
        reflect.SliceOf(fnVal.Type().Out(0)),
        sliceVal.Len(),
        sliceVal.Len(),
    )

    for i := 0; i < sliceVal.Len(); i++ {
        result := fnVal.Call([]reflect.Value{sliceVal.Index(i)})
        resultSlice.Index(i).Set(result[0])
    }

    return resultSlice.Interface()
}

// Usage
// doubled := Map([]int{1, 2, 3}, func(x int) int { return x * 2 }).([]int)
```

### 2. Unsafe

```go
import "unsafe"

// Unsafe pointer operations
// USE WITH EXTREME CAUTION

// Size of types
size := unsafe.Sizeof(int(0))     // 8 on 64-bit
align := unsafe.Alignof(struct{}{})

// Pointer arithmetic (very dangerous)
func unsafeSliceAccess(data []byte, offset int) byte {
    ptr := unsafe.Pointer(&data[0])
    ptr = unsafe.Pointer(uintptr(ptr) + uintptr(offset))
    return *(*byte)(ptr)
}

// Converting between types (unsafe)
func bytesToString(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}

func stringToBytes(s string) []byte {
    return *(*[]byte)(unsafe.Pointer(
        &struct {
            string
            int
        }{s, len(s)},
    ))
}

// Access unexported fields (DON'T DO THIS IN PRODUCTION)
func getUnexportedField(obj interface{}, field string) interface{} {
    v := reflect.ValueOf(obj).Elem().FieldByName(field)
    return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Interface()
}

// When to use unsafe:
// 1. Performance-critical code (measure first!)
// 2. Interfacing with C code
// 3. Implementing low-level libraries
//
// When NOT to use unsafe:
// 1. Normal application code
// 2. When there's a safe alternative
// 3. Without thorough testing
```

### 3. CGO

```go
/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

void printHello() {
    printf("Hello from C!\n");
}

int add(int a, int b) {
    return a + b;
}

char* duplicate(const char* s) {
    char* copy = malloc(strlen(s) + 1);
    strcpy(copy, s);
    return copy;
}
*/
import "C"
import (
    "fmt"
    "unsafe"
)

func main() {
    // Call C function
    C.printHello()

    // Pass integers
    result := C.add(1, 2)
    fmt.Println(result)

    // Pass strings
    goStr := "hello"
    cStr := C.CString(goStr)        // Allocates in C
    defer C.free(unsafe.Pointer(cStr))  // Must free!

    copy := C.duplicate(cStr)
    defer C.free(unsafe.Pointer(copy))

    // Convert C string to Go string
    goCopy := C.GoString(copy)
    fmt.Println(goCopy)

    // C array to Go slice
    cArray := C.malloc(C.size_t(10) * C.size_t(unsafe.Sizeof(C.int(0))))
    defer C.free(cArray)

    goSlice := (*[10]C.int)(cArray)[:]
    goSlice[0] = 42
}

// Building with cgo
// go build
// CGO_ENABLED=0 go build  # Disable cgo

// Using external C libraries
/*
#cgo pkg-config: openssl
#include <openssl/ssl.h>
*/
import "C"
```

### 4. Performance Optimization

```go
// 1. Avoid allocations in hot paths

// Bad
func processItems(items []string) []string {
    var result []string
    for _, item := range items {
        result = append(result, strings.ToUpper(item))
    }
    return result
}

// Good
func processItems(items []string) []string {
    result := make([]string, 0, len(items))  // Pre-allocate
    for _, item := range items {
        result = append(result, strings.ToUpper(item))
    }
    return result
}

// 2. Use sync.Pool for reusable objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processWithPool() string {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()

    buf.WriteString("processed")
    return buf.String()
}

// 3. Avoid interface{} when possible

// Bad
func sum(numbers []interface{}) int64 {
    var total int64
    for _, n := range numbers {
        total += n.(int64)
    }
    return total
}

// Good - use generics
func sum[T constraints.Integer](numbers []T) T {
    var total T
    for _, n := range numbers {
        total += n
    }
    return total
}

// 4. String builder for concatenation

// Bad
func concat(parts []string) string {
    var result string
    for _, part := range parts {
        result += part  // Creates new string each time
    }
    return result
}

// Good
func concat(parts []string) string {
    var sb strings.Builder
    sb.Grow(len(parts) * 20)  // Estimate size
    for _, part := range parts {
        sb.WriteString(part)
    }
    return sb.String()
}

// 5. Avoid locks in hot paths with atomic operations

// Bad
var counter int
var mu sync.Mutex

func increment() {
    mu.Lock()
    counter++
    mu.Unlock()
}

// Good
var counter int64

func increment() {
    atomic.AddInt64(&counter, 1)
}

// 6. Batch operations
func batchInsert(db *sql.DB, users []User) error {
    tx, _ := db.Begin()
    stmt, _ := tx.Prepare("INSERT INTO users (name, email) VALUES ($1, $2)")
    defer stmt.Close()

    for _, user := range users {
        stmt.Exec(user.Name, user.Email)
    }

    return tx.Commit()
}

// 7. Use appropriate data structures
// - Map for O(1) lookups
// - Slice for sequential access
// - Heap for priority queues
// - Ring buffer for fixed-size queues

// 8. Escape analysis
// Run: go build -gcflags='-m'

// Keep values on stack when possible
func noEscape() int {
    x := 42
    return x  // Copied, stays on stack
}

func escapes() *int {
    x := 42
    return &x  // Escapes to heap
}
```

### 5. Memory Management

```go
// Set memory limit (Go 1.19+)
import "runtime/debug"

func init() {
    debug.SetMemoryLimit(1 * GiB)
}

// GC tuning
import "runtime"

// Force GC (rarely needed)
runtime.GC()

// GC percentage (default 100)
debug.SetGCPercent(100)  // GC when heap grows 100%

// Read GC stats
var stats debug.GCStats
debug.ReadGCStats(&stats)

// Memory stats
var m runtime.MemStats
runtime.ReadMemStats(&m)
fmt.Printf("Alloc = %v MiB\n", m.Alloc/1024/1024)
fmt.Printf("TotalAlloc = %v MiB\n", m.TotalAlloc/1024/1024)
fmt.Printf("Sys = %v MiB\n", m.Sys/1024/1024)
fmt.Printf("NumGC = %v\n", m.NumGC)

// Finalizers (use sparingly)
func NewResource() *Resource {
    r := &Resource{}
    runtime.SetFinalizer(r, (*Resource).Close)
    return r
}
```

### 6. Best Practices

```go
// 1. Project Layout
/*
myapp/
├── cmd/                    # Main applications
│   ├── server/
│   │   └── main.go
│   └── cli/
│       └── main.go
├── internal/               # Private application code
│   ├── handler/
│   ├── service/
│   └── repository/
├── pkg/                    # Public library code
├── api/                    # API definitions
├── configs/                # Configuration files
├── scripts/                # Build and deployment scripts
├── deployments/            # Docker, K8s configs
├── docs/                   # Documentation
├── go.mod
├── go.sum
├── Makefile
└── README.md
*/

// 2. Configuration
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
}

func LoadConfig() (*Config, error) {
    // Use environment variables with defaults
    return &Config{
        Server: ServerConfig{
            Port: getEnv("SERVER_PORT", "8080"),
        },
    }, nil
}

// 3. Graceful shutdown
func main() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    server := &http.Server{Addr: ":8080"}

    go func() {
        <-ctx.Done()
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        server.Shutdown(shutdownCtx)
    }()

    server.ListenAndServe()
}

// 4. Structured logging
import "log/slog"

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    slog.SetDefault(logger)

    slog.Info("server starting",
        "port", 8080,
        "version", "1.0.0",
    )
}

// 5. Health checks
func healthHandler(w http.ResponseWriter, r *http.Request) {
    // Check dependencies
    if err := db.Ping(); err != nil {
        http.Error(w, "database unavailable", http.StatusServiceUnavailable)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// 6. Metrics
import "github.com/prometheus/client_golang/prometheus"

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
}

// 7. Error handling in production
func errorHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                slog.Error("panic recovered",
                    "error", err,
                    "path", r.URL.Path,
                    "stack", string(debug.Stack()),
                )
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// 8. Context propagation
func (s *Service) Process(ctx context.Context, id int) error {
    // Add tracing
    span, ctx := opentracing.StartSpanFromContext(ctx, "Service.Process")
    defer span.Finish()

    // Add logging context
    logger := slog.With("request_id", ctx.Value("requestID"))

    // Propagate to all operations
    return s.repo.Get(ctx, id)
}
```

### 7. Common Patterns

```go
// 1. Repository pattern
type UserRepository interface {
    Get(ctx context.Context, id int64) (*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id int64) error
}

// 2. Service layer
type UserService struct {
    repo   UserRepository
    cache  Cache
    events EventEmitter
}

func (s *UserService) Get(ctx context.Context, id int64) (*User, error) {
    // Check cache
    if user, err := s.cache.Get(ctx, id); err == nil {
        return user, nil
    }

    // Get from repository
    user, err := s.repo.Get(ctx, id)
    if err != nil {
        return nil, err
    }

    // Cache result
    s.cache.Set(ctx, user)

    return user, nil
}

// 3. Dependency injection with Wire
// wire.go
//go:build wireinject

func InitializeServer() (*Server, error) {
    wire.Build(
        NewConfig,
        NewDatabase,
        NewRedis,
        NewUserRepository,
        NewUserService,
        NewUserHandler,
        NewServer,
    )
    return nil, nil
}

// 4. Clean architecture layers
/*
Domain (entities) -> Use Cases (services) -> Interface Adapters (handlers) -> Frameworks & Drivers (db, http)
*/

// 5. Event-driven architecture
type EventEmitter interface {
    Emit(ctx context.Context, event Event) error
    Subscribe(eventType string, handler EventHandler) error
}

func (s *UserService) Create(ctx context.Context, user *User) error {
    if err := s.repo.Create(ctx, user); err != nil {
        return err
    }

    // Emit event
    s.events.Emit(ctx, Event{
        Type: "user.created",
        Data: user,
    })

    return nil
}
```

## Hands-on Exercises

Create the following programs in `phase-12-advanced/`:

### Exercise 1: Reflection-Based Validator

Build a validation library using reflection:

- Tag-based validation rules
- Nested struct support
- Custom validators
- Error messages

### Exercise 2: Performance Optimization

Take a slow program and optimize it:

- Profile and identify bottlenecks
- Reduce allocations
- Improve concurrency
- Document improvements

### Exercise 3: CGO Integration

Create a Go wrapper for a C library:

- Memory management
- Error handling
- Thread safety
- Tests

### Exercise 4: Production Service

Build a production-ready service with:

- Graceful shutdown
- Health checks
- Metrics
- Structured logging
- Configuration management

## Resources

### Official

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Blog](https://go.dev/blog/)

### Books

- [The Go Programming Language](https://www.gopl.io/)
- [Concurrency in Go](https://www.oreilly.com/library/view/concurrency-in-go/9781491941294/)
- [100 Go Mistakes and How to Avoid Them](https://100go.co/)

### Advanced

- [Go Internals](https://github.com/teh-cmc/go-internals)
- [Go Garbage Collector](https://tip.golang.org/doc/gc-guide)

## Validation Checklist

- [ ] Understand reflection basics
- [ ] Know when to use unsafe
- [ ] Can use cgo for C integration
- [ ] Can profile and optimize code
- [ ] Apply production best practices
- [ ] All exercises completed

## Congratulations!

You've completed the Go learning path! You now have the knowledge to:

- Write idiomatic Go code
- Build concurrent applications
- Create web services and APIs
- Work with databases
- Write comprehensive tests
- Use Go tooling effectively
- Optimize for performance
- Apply production best practices

### Next Steps

1. **Build Real Projects**: Apply your knowledge to real-world problems
2. **Contribute to Open Source**: Find Go projects on GitHub
3. **Read Go Source Code**: Learn from the standard library
4. **Join the Community**:
   - [Go Forum](https://forum.golangbridge.org/)
   - [Gophers Slack](https://gophers.slack.com/)
   - [r/golang](https://reddit.com/r/golang)
5. **Continue Learning**: Go is constantly evolving - stay updated!

Happy coding! 🎉
