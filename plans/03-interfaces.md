# Phase 03: Interfaces and Type System

**Duration**: 1-2 weeks  
**Prerequisites**: Phase 02 completed  
**Practice Directory**: `phase-03-interfaces/`

## Overview

Go's interfaces are fundamentally different from classes in TypeScript/PHP/Python. They're implicit, composable, and enable duck typing with compile-time safety. This phase covers Go's type system, interfaces, type assertions, and generics (introduced in Go 1.18).

## Learning Objectives

- Understand implicit interface satisfaction
- Master interface composition and embedding
- Use type assertions and type switches safely
- Learn when and how to use the empty interface
- Understand Go 1.18+ generics
- Apply interface best practices

## Topics to Cover

### 1. Interface Basics

```go
// Interface definition
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// Implicit satisfaction - no "implements" keyword!
type File struct {
    // ...
}

func (f *File) Read(p []byte) (n int, err error) {
    // Implementation
    return 0, nil
}

// File now satisfies Reader - compiler knows automatically

// Using the interface
func Process(r Reader) {
    buf := make([]byte, 1024)
    r.Read(buf)
}

func main() {
    f := &File{}
    Process(f)  // File satisfies Reader
}
```

**Key Insight**: Interfaces are satisfied implicitly. If your type has the right methods, it implements the interface. No explicit declaration needed.

### 2. Interface Composition

```go
// Combining interfaces
type ReadWriter interface {
    Reader
    Writer
}

type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}

// Standard library example
type Closer interface {
    Close() error
}

// io package has many composed interfaces
// io.Reader, io.Writer, io.ReadWriter, io.ReadCloser, etc.
```

### 3. Interface Values

```go
// Interface value has two parts: (type, value)
var r Reader
fmt.Printf("%T, %v\n", r, r)  // <nil>, <nil>

r = &File{}
fmt.Printf("%T, %v\n", r, r)  // *File, &{...}

// Interface with nil underlying value
var f *File  // nil
r = f        // r is NOT nil! It has type *File
fmt.Printf("%T, %v\n", r, r)  // *File, <nil>

// This is a common source of bugs!
func isNil(r Reader) bool {
    return r == nil  // Wrong! Check below
}

func isNilCorrect(r Reader) bool {
    if r == nil {
        return true
    }
    v := reflect.ValueOf(r)
    switch v.Kind() {
    case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
        return v.IsNil()
    }
    return false
}
```

**The Nil Interface Trap**:

```go
func returnsError() error {
    var err *MyError  // nil
    return err        // Returns (type=*MyError, value=nil), NOT nil!
}

func main() {
    if returnsError() != nil {  // This is TRUE!
        fmt.Println("Error!")   // Prints even though err is nil
    }
}

// Correct approach
func returnsErrorCorrect() error {
    var err *MyError
    if err != nil {
        return err
    }
    return nil  // Return nil interface, not nil typed pointer
}
```

### 4. Type Assertions and Type Switches

```go
// Type assertion
var r Reader = &File{}

f, ok := r.(*File)  // Safe assertion with ok
if ok {
    // f is *File
}

f := r.(*File)      // Unsafe - panics if wrong type

// Type switch
func describe(v interface{}) {
    switch v := v.(type) {
    case int:
        fmt.Printf("int: %d\n", v)
    case string:
        fmt.Printf("string: %s\n", v)
    case []int:
        fmt.Printf("[]int: %v\n", v)
    case io.Reader:
        fmt.Printf("Reader: %T\n", v)
    default:
        fmt.Printf("unknown: %T\n", v)
    }
}

// Assertion on interface
func process(r Reader) {
    if rw, ok := r.(io.ReadWriter); ok {
        // r also implements ReadWriter
        rw.Write([]byte("hello"))
    }
}
```

### 5. Empty Interface

```go
// interface{} - can hold any value
var any interface{}

any = 42
any = "hello"
any = []int{1, 2, 3}
any = func() {}

// Go 1.18+ alias
var any2 any  // same as interface{}

// Use cases
func Print(v any) {
    fmt.Println(v)
}

// JSON handling
var data map[string]interface{}
json.Unmarshal(jsonBytes, &data)

// But prefer typed structures when possible!
type Response struct {
    Status  string `json:"status"`
    Message string `json:"message"`
}
```

**When to Use Empty Interface**:

- JSON parsing of unknown structure
- Generic containers before generics
- fmt.Println-style functions
- Avoid when possible - lose type safety

### 6. Generics (Go 1.18+)

```go
// Generic function
func Min[T constraints.Ordered](a, b T) T {
    if a < b {
        return a
    }
    return b
}

// Usage
minInt := Min(1, 2)
minFloat := Min(1.5, 2.5)
minString := Min("a", "b")

// Generic type
type Stack[T any] struct {
    elements []T
}

func (s *Stack[T]) Push(v T) {
    s.elements = append(s.elements, v)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.elements) == 0 {
        var zero T
        return zero, false
    }
    v := s.elements[len(s.elements)-1]
    s.elements = s.elements[:len(s.elements)-1]
    return v, true
}

// Usage
intStack := Stack[int]{}
intStack.Push(1)

stringStack := Stack[string]{}
stringStack.Push("hello")

// Type constraints
type Number interface {
    int | int64 | float64
}

func Sum[T Number](nums []T) T {
    var total T
    for _, n := range nums {
        total += n
    }
    return total
}

// Custom constraints
type Comparable[T any] interface {
    CompareTo(T) int
}

// Constraint from constraints package
import "golang.org/x/exp/constraints"

func Max[T constraints.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}

// Generic with multiple type parameters
func Pair[K any, V any](k K, v V) Pair[K, V] {
    return Pair[K, V]{Key: k, Value: v}
}

// Type inference
// Go can often infer type arguments
result := Min(1, 2)  // T inferred as int
```

**Generics Best Practices**:

```go
// When to use generics:
// 1. Container types (Stack, Queue, List)
// 2. Algorithms that work on multiple types
// 3. When you'd otherwise use interface{} and type assertions

// When NOT to use generics:
// 1. When interfaces work fine
// 2. For single-method interfaces (keep simple)
// 3. When it makes code harder to read

// Prefer interfaces for behavior
type Stringer interface {
    String() string
}

func PrintAll[T Stringer](items []T) {  // Overkill
    for _, item := range items {
        fmt.Println(item.String())
    }
}

func PrintAllSimple(items []Stringer) {  // Better
    for _, item := range items {
        fmt.Println(item.String())
    }
}
```

### 7. Interface Design Patterns

```go
// 1. Small interfaces (Go idiom)
// Good: Single method
type Reader interface {
    Read(p []byte) (n int, err error)
}

// Bad: Many methods
type MonsterInterface interface {
    Read() error
    Write() error
    Close() error
    Flush() error
    Reset() error
    // ... 10 more methods
}

// 2. Accept interfaces, return structs
func NewClient(cfg Config) *Client {  // Return concrete type
    return &Client{config: cfg}
}

func (c *Client) Do(req Request) (Response, error) {
    // ...
}

// Users can mock Client if needed
type Doer interface {
    Do(req Request) (Response, error)
}

func Process(d Doer) {  // Accept interface
    d.Do(Request{})
}

// 3. Interface for testing
type Database interface {
    GetUser(id int) (*User, error)
    SaveUser(u *User) error
}

type RealDatabase struct { /* ... */ }
type MockDatabase struct { /* ... */ }

// 4. Decorator pattern
type LoggingReader struct {
    r io.Reader
}

func (lr *LoggingReader) Read(p []byte) (n int, err error) {
    n, err = lr.r.Read(p)
    log.Printf("read %d bytes", n)
    return
}

// 5. Adapter pattern
type LegacyWriter struct {
    // ...
}

func (lw *LegacyWriter) Write(data string) error {
    // Old interface
    return nil
}

// Adapter to new interface
type WriterAdapter struct {
    legacy *LegacyWriter
}

func (wa *WriterAdapter) Write(p []byte) (n int, err error) {
    err = wa.legacy.Write(string(p))
    return len(p), err
}
```

### 8. Practical Examples

```go
// Repository pattern
type Repository[T any] interface {
    Get(id string) (T, error)
    Save(entity T) error
    Delete(id string) error
    List() ([]T, error)
}

type UserRepository struct {
    db *sql.DB
}

func (r *UserRepository) Get(id string) (*User, error) {
    // Implementation
    return nil, nil
}

// Middleware pattern
type Handler func(w http.ResponseWriter, r *http.Request)

func LoggingMiddleware(next Handler) Handler {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next(w, r)
        log.Printf("%s %s took %v", r.Method, r.URL, time.Since(start))
    }
}

// Chain middleware
func Chain(h Handler, middlewares ...func(Handler) Handler) Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}
```

## Hands-on Exercises

Create the following programs in `phase-03-interfaces/`:

### Exercise 1: Generic Collections

Implement generic Stack, Queue, and Set types with full test coverage.

### Exercise 2: Sort Package Recreation

Create a generic sort function that works with any ordered type.

### Exercise 3: Mock Testing

Create an interface for a weather service, implement a real and mock version, and write tests using both.

### Exercise 4: Plugin System

Design a plugin system using interfaces where plugins can be loaded dynamically.

## Resources

### Official

- [Effective Go: Interfaces](https://go.dev/doc/effective_go#interfaces)
- [Go Generics Tutorial](https://go.dev/doc/tutorial/generics)
- [The Laws of Reflection](https://go.dev/blog/laws-of-reflection)

### Deep Dives

- [Go Data Structures: Interfaces](https://research.swtch.com/interfaces)
- [When To Use Generics](https://go.dev/blog/when-generics)

## Validation Checklist

- [ ] Understand implicit interface satisfaction
- [ ] Can compose interfaces
- [ ] Understand interface value representation (type, value)
- [ ] Know the nil interface trap
- [ ] Can use type assertions safely
- [ ] Understand when to use empty interface
- [ ] Can write generic functions and types
- [ ] Know when to use generics vs interfaces
- [ ] All exercises completed

## Common Pitfalls

| Mistake                                  | Consequence       | Fix                            |
| ---------------------------------------- | ----------------- | ------------------------------ |
| Returning nil typed pointer as interface | Non-nil interface | Return nil directly            |
| Large interfaces                         | Hard to implement | Keep interfaces small          |
| Overusing empty interface                | Lose type safety  | Use generics or specific types |
| Premature generics                       | Complex code      | Start with interfaces          |

## Next Phase

Proceed to **Phase 04: Error Handling and Idioms** to master Go's unique approach to error handling.
