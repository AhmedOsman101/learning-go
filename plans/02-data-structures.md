# Phase 02: Data Structures and Memory

**Duration**: 1-2 weeks  
**Prerequisites**: Phase 01 completed  
**Practice Directory**: `phase-02-data-structures/`

## Overview

This phase dives deep into Go's core data structures: arrays, slices, maps, and structs. Understanding how these work internally is crucial for writing efficient Go code. We'll also cover Go's memory model and when to use pointers vs values.

## Learning Objectives

- Master arrays vs slices and their internal representation
- Understand slice capacity vs length
- Work with maps and understand their hash table implementation
- Define and use structs effectively
- Understand value vs pointer semantics
- Learn Go's memory allocation patterns

## Topics to Cover

### 1. Arrays

```go
// Arrays have fixed size - part of the type!
var arr1 [5]int                    // [0 0 0 0 0]
arr2 := [5]int{1, 2, 3, 4, 5}      // Initialized
arr3 := [...]int{1, 2, 3}          // Size inferred: [3]int
arr4 := [5]int{1: 10, 3: 30}       // [0 10 0 30 0]

// Arrays are values - copied on assignment!
arr5 := arr2      // Copy of entire array
arr5[0] = 100     // Doesn't affect arr2

// Passing to function - copies the array
func modifyArray(arr [5]int) {
    arr[0] = 999  // Only modifies local copy
}

// Pass pointer to avoid copy
func modifyArrayPtr(arr *[5]int) {
    arr[0] = 999  // Modifies original
}

// Multidimensional
var matrix [3][3]int
```

**Key Insight**: Arrays are rarely used directly in Go. Slices are the go-to choice for dynamic collections.

### 2. Slices - The Heart of Go Collections

```go
// Slices are views into arrays
slice1 := []int{1, 2, 3, 4, 5}     // Slice literal
slice2 := make([]int, 5)           // len=5, cap=5
slice3 := make([]int, 5, 10)       // len=5, cap=10
slice4 := new([10]int)[:5]         // Rarely used

// Slice internals
type slice struct {
    ptr *[T]    // pointer to underlying array
    len int     // number of elements
    cap int     // capacity of underlying array
}

// Length and capacity
fmt.Println(len(slice1), cap(slice1))

// Slicing operations
arr := [10]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
s1 := arr[2:5]    // [2 3 4], len=3, cap=8
s2 := arr[:5]     // [0 1 2 3 4], len=5, cap=10
s3 := arr[5:]     // [5 6 7 8 9], len=5, cap=5
s4 := arr[:]      // Full slice

// Reslicing (within capacity)
s5 := s1[1:4]     // [3 4 5], extends beyond original slice

// Append - may reallocate
s6 := append(s1, 10, 11)  // Returns new slice

// Common pattern: append to nil slice
var s []int           // nil slice
s = append(s, 1, 2)   // Works! Creates underlying array

// Copy slices
src := []int{1, 2, 3}
dst := make([]int, len(src))
copy(dst, src)

// Delete from slice
slice := []int{1, 2, 3, 4, 5}
// Delete index 2
slice = append(slice[:2], slice[3:]...)

// Insert at index
slice = append(slice[:2], append([]int{99}, slice[2:]...)...)

// Filter in place
slice = slice[:0]  // Clear but keep capacity
for _, v := range data {
    if condition(v) {
        slice = append(slice, v)
    }
}
```

**Memory Layout**:

```
Array:  [0][1][2][3][4][5][6][7][8][9]
Slice:      [2][3][4]  <- points to array[2]
            len=3, cap=8
```

**Key Differences from Other Languages**:

- Slices are views, not copies
- Appending may create new underlying array
- Slices share memory with original array

### 3. Maps

```go
// Map creation
m1 := make(map[string]int)
m2 := map[string]int{"a": 1, "b": 2}
m3 := make(map[string]int, 100)  // Hint initial size

// Operations
m1["key"] = 100           // Insert/update
value := m1["key"]        // Get (zero value if missing)
value, exists := m1["key"] // Get with existence check
delete(m1, "key")         // Delete

// Iteration (random order!)
for key, value := range m2 {
    fmt.Printf("%s: %d\n", key, value)
}

// Keys only
for key := range m2 {
    fmt.Println(key)
}

// Map of slices
m := make(map[string][]int)
m["numbers"] = []int{1, 2, 3}

// Map of maps
nested := make(map[string]map[string]int)
nested["outer"] = map[string]int{"inner": 1}

// Thread-safe map (use sync.Map or mutex)
var mu sync.Mutex
m := make(map[string]int)
mu.Lock()
m["key"] = 1
mu.Unlock()
```

**Map Internals**:

- Hash table implementation
- Buckets hold 8 key-value pairs
- Automatic growing/shrinking
- Random iteration order by design

**Common Patterns**:

```go
// Set implementation (map with empty struct)
set := make(map[string]struct{})
set["item"] = struct{}{}
if _, exists := set["item"]; exists {
    // Item in set
}

// Counting
counts := make(map[string]int)
for _, word := range words {
    counts[word]++
}

// Grouping
groups := make(map[string][]Item)
for _, item := range items {
    groups[item.Category] = append(groups[item.Category], item)
}
```

### 4. Structs

```go
// Basic struct
type Person struct {
    Name string
    Age  int
    City string
}

// Creation
p1 := Person{Name: "John", Age: 30}
p2 := Person{"Jane", 25, "NYC"}        // Must specify all fields
p3 := new(Person)                       // Returns *Person, all zero values
p4 := &Person{Name: "Bob"}              // Pointer literal

// Accessing
fmt.Println(p1.Name)
p1.Age = 31

// Methods on structs
func (p Person) String() string {
    return fmt.Sprintf("%s (%d)", p.Name, p.Age)
}

// Value receiver - doesn't modify original
func (p Person) SetAgeValue(age int) {
    p.Age = age  // Only modifies copy
}

// Pointer receiver - modifies original
func (p *Person) SetAge(age int) {
    p.Age = age  // Modifies original
}

// Embedding (composition, not inheritance)
type Employee struct {
    Person      // Embedded (anonymous field)
    Title string
    Salary float64
}

e := Employee{
    Person: Person{Name: "John", Age: 30},
    Title:  "Engineer",
}
fmt.Println(e.Name)      // Access embedded field directly
fmt.Println(e.Person.Name) // Or explicitly

// Tags (for JSON, DB, etc.)
type User struct {
    ID       int    `json:"id" db:"user_id"`
    Name     string `json:"name" validate:"required"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"-"` // Omit from JSON
}

// Struct comparison (if all fields comparable)
p1 := Person{Name: "John", Age: 30}
p2 := Person{Name: "John", Age: 30}
fmt.Println(p1 == p2)  // true
```

**When to Use Pointer vs Value Receiver**:

```go
// Use pointer receiver when:
// 1. Method needs to modify the receiver
// 2. Struct is large (avoid copy)
// 3. Consistency (if some methods need pointer, all should use pointer)

// Use value receiver when:
// 1. Struct is small
// 2. Method is read-only
// 3. You want value semantics
```

### 5. Memory Allocation

```go
// Stack vs Heap
// Go compiler decides - escape analysis

// Likely on stack
func stackAlloc() int {
    x := 42
    return x  // Returns copy
}

// Likely on heap (escapes to heap)
func heapAlloc() *int {
    x := 42
    return &x  // Returns pointer, x must survive
}

// View escape analysis
// go build -gcflags='-m'

// Common allocations
// make() for slices, maps, channels
// new() for pointers to zero values (rarely used)

// Slice allocation patterns
// Good: pre-allocate if size known
result := make([]int, 0, expectedSize)
for i := 0; i < expectedSize; i++ {
    result = append(result, compute(i))
}

// Avoid: repeated allocations in loop
var result []int
for i := 0; i < 1000; i++ {
    result = append(result, compute(i))  // May reallocate multiple times
}
```

### 6. Practical Patterns

```go
// Builder pattern for complex structs
type Request struct {
    URL     string
    Method  string
    Headers map[string]string
    Body    []byte
}

type RequestBuilder struct {
    request Request
}

func NewRequestBuilder(url string) *RequestBuilder {
    return &RequestBuilder{
        request: Request{URL: url, Method: "GET"},
    }
}

func (b *RequestBuilder) Method(method string) *RequestBuilder {
    b.request.Method = method
    return b
}

func (b *RequestBuilder) Header(key, value string) *RequestBuilder {
    if b.request.Headers == nil {
        b.request.Headers = make(map[string]string)
    }
    b.request.Headers[key] = value
    return b
}

func (b *RequestBuilder) Build() Request {
    return b.request
}

// Usage
req := NewRequestBuilder("https://api.example.com").
    Method("POST").
    Header("Content-Type", "application/json").
    Build()

// Functional options pattern
type Server struct {
    host string
    port int
    tls  bool
}

type Option func(*Server)

func WithHost(host string) Option {
    return func(s *Server) { s.host = host }
}

func WithPort(port int) Option {
    return func(s *Server) { s.port = port }
}

func WithTLS() Option {
    return func(s *Server) { s.tls = true }
}

func NewServer(opts ...Option) *Server {
    s := &Server{host: "localhost", port: 8080}
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Usage
server := NewServer(WithPort(3000), WithTLS())
```

## Hands-on Exercises

Create the following programs in `phase-02-data-structures/`:

### Exercise 1: Slice Operations Library

Create a package with generic slice operations:

- Filter, Map, Reduce
- Chunk, Flatten
- Unique, Intersect

### Exercise 2: Custom HashMap

Implement a simple hash map from scratch using slices and understand collision handling.

### Exercise 3: JSON Parser

Parse a JSON file into structs using tags. Handle nested structures.

### Exercise 4: Memory Profiling

Write a program that demonstrates:

- Stack vs heap allocation
- Escape analysis
- Memory leaks with slices

## Resources

### Official

- [Go Slices: usage and internals](https://go.dev/blog/slices)
- [Go Maps in Action](https://go.dev/blog/maps)
- [Allocation efficiency in Go](https://go.dev/doc/faq#stack_or_heap)

### Deep Dives

- [SliceTricks](https://github.com/golang/go/wiki/SliceTricks) - Common slice operations
- [Understanding Slice Capacity](https://www.ardanlabs.com/blog/2013/08/understanding-slices-in-go-programming.html)

## Validation Checklist

- [ ] Understand array vs slice difference
- [ ] Can explain slice length vs capacity
- [ ] Know when append creates new underlying array
- [ ] Can use maps with existence checking
- [ ] Understand struct embedding
- [ ] Know when to use pointer vs value receiver
- [ ] Understand escape analysis basics
- [ ] All exercises completed

## Common Pitfalls

| Mistake                         | Consequence             | Fix                         |
| ------------------------------- | ----------------------- | --------------------------- |
| `s = s[:cap(s)]`                | May access garbage data | Check bounds carefully      |
| Modifying slice while iterating | Unexpected behavior     | Use separate slice or index |
| Map concurrent access           | Race condition          | Use mutex or sync.Map       |
| Large struct copies             | Performance hit         | Use pointer                 |

## Next Phase

Proceed to **Phase 03: Interfaces and Type System** to learn Go's approach to polymorphism and type assertions.
