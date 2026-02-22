# Phase 01: Go Fundamentals for Experienced Developers

**Duration**: 1-2 weeks  
**Prerequisites**: Programming experience (TypeScript, PHP, Python, Rust)  
**Practice Directory**: `phase-01-fundamentals/`

## Overview

This phase focuses on Go-specific syntax and tooling, skipping general programming concepts you already know. We'll emphasize Go's unique characteristics and how they differ from languages you're familiar with.

## Learning Objectives

- Set up Go development environment
- Understand Go's project structure and module system
- Master Go syntax differences from TypeScript/Python/Rust
- Learn Go's type system and declarations
- Understand Go's approach to package management

## Topics to Cover

### 1. Installation and Tooling

```bash
# Verify installation
go version
go env

# Essential commands
go mod init <module-name>    # Initialize module (like package.json)
go mod tidy                   # Clean up dependencies
go get <package>              # Add dependency
go build                      # Compile
go run <file>                 # Compile and run
go test                       # Run tests
go fmt                        # Format code (opinionated)
go vet                        # Static analysis
```

**Key Differences from Other Languages**:

- No package managers like npm/pip/composer - Go has built-in module system
- `go fmt` is non-negotiable - no prettier config debates
- `go vet` catches common mistakes before runtime

### 2. Project Structure

```
my-project/
├── go.mod              # Module definition (like package.json)
├── go.sum              # Dependency checksums (like package-lock.json)
├── main.go             # Entry point (must have main package + main func)
├── internal/           # Private application code
│   └── handlers/
├── pkg/                # Public library code
├── cmd/                # Multiple executables
│   └── myapp/
│       └── main.go
└── api/                # API definitions (OpenAPI, protobuf)
```

**Key Insight**: Go favors flat structures over deep nesting. Don't over-organize.

### 3. Variable Declarations and Types

```go
// Variable declarations - multiple styles
var name string = "John"           // Explicit type
var name = "John"                  // Type inference
name := "John"                     // Short declaration (most common)

// Constants
const Pi = 3.14159
const (
    StatusOK    = 200
    StatusError = 500
)

// Basic types
var (
    i   int       = 42
    f   float64   = 3.14
    b   bool      = true
    s   string    = "hello"
    r   rune      = 'A'        // Unicode code point (int32)
    by  byte      = 0xFF       // uint8
)

// Zero values (no undefined/null!)
var (
    zeroInt    int     // 0
    zeroFloat  float64 // 0.0
    zeroBool   bool    // false
    zeroString string  // "" (empty string, not null!)
    zeroPtr    *int    // nil
)
```

**Key Differences**:

- No `undefined` or `null` - everything has a zero value
- `:=` is your friend - use it inside functions
- `var` for package-level variables or when you need zero value explicitly

### 4. Functions

```go
// Basic function
func add(a, b int) int {
    return a + b
}

// Multiple return values (Go's signature feature)
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// Named return values (use sparingly)
func calculate(a, b int) (sum, product int) {
    sum = a + b
    product = a * b
    return // naked return - implicit return of named values
}

// Variadic functions
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

// Function as value (like TypeScript)
func apply(nums []int, fn func(int) int) []int {
    result := make([]int, len(nums))
    for i, n := range nums {
        result[i] = fn(n)
    }
    return result
}

// Closures
func counter() func() int {
    count := 0
    return func() int {
        count++
        return count
    }
}
```

**Key Differences from TypeScript/Python**:

- Multiple return values are idiomatic (not tuples)
- No default parameters - use functional options pattern instead
- No function overloading

### 5. Control Flow

```go
// If statements - no parentheses required
if err != nil {
    return err
}

// If with initialization (like Rust's let-if)
if err := doSomething(); err != nil {
    return err
}
// err is scoped to if-else block

// For loop - Go only has for (no while, no do-while)
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// While-like
for condition {
    // do something
}

// Infinite loop
for {
    if shouldBreak {
        break
    }
}

// Range-based (like for...of in JS)
for index, value := range slice {
    fmt.Printf("%d: %v\n", index, value)
}

// Ignore index
for _, value := range slice {
    fmt.Println(value)
}

// Switch - no break needed (auto-break)
switch value {
case 1:
    fmt.Println("one")
case 2, 3:  // multiple cases
    fmt.Println("two or three")
default:
    fmt.Println("other")
}

// Switch without condition (like if-else chain)
switch {
case score >= 90:
    grade = "A"
case score >= 80:
    grade = "B"
default:
    grade = "F"
}

// Fallthrough (rarely used)
switch num {
case 1:
    fmt.Println("one")
    fallthrough
case 2:
    fmt.Println("two")  // will execute even if num is 1
}
```

**Key Differences**:

- Only `for` loop - no `while`, `do-while`, `for...in`, `for...of`
- `switch` breaks automatically - use `fallthrough` explicitly
- `if` can have initialization statement (similar to Rust)

### 6. Packages and Imports

```go
// main.go
package main  // Executable must be package main

import (
    "fmt"           // Standard library
    "strings"       // Standard library

    "github.com/user/project/pkg/handler"  // External package
    . "github.com/user/project/pkg/utils"  // Dot import (avoid!)
    alias "github.com/user/project/pkg/longname"  // Aliased import
)

// Internal package (not importable from outside module)
import "myproject/internal/database"

// Multiple imports grouped
import (
    "fmt"
    "os"
)
```

**Visibility Rules**:

```go
package mypackage

// Public (exported) - starts with uppercase
func PublicFunction() {}

// Private (unexported) - starts with lowercase
func privateFunction() {}

type MyStruct struct {
    PublicField   int   // accessible from outside
    privateField  int   // only accessible within package
}
```

**Key Insight**: Go uses capitalization for visibility, not `public`/`private` keywords.

### 7. Pointers (if coming from TypeScript/Python)

```go
// Go has pointers, but they're simpler than C/C++
func increment(n *int) {
    *n++  // dereference and modify
}

func main() {
    x := 5
    increment(&x)  // pass address
    fmt.Println(x) // 6

    // Pointer to struct
    p := &Person{Name: "John"}
    p.Name = "Jane"  // No need to dereference for struct fields

    // Nil pointer
    var ptr *int  // nil
    if ptr != nil {
        *ptr = 10  // Safe access
    }
}
```

**Key Differences from Rust**:

- No ownership/borrowing rules - simpler but less safe
- No null safety - you must check for nil
- Garbage collected - no manual memory management

## Hands-on Exercises

Create the following programs in `phase-01-fundamentals/`:

### Exercise 1: Hello Modules

```bash
# In phase-01-fundamentals/exercise-01/
go mod init github.com/yourusername/learning/phase-01-fundamentals/exercise-01
```

Create a simple CLI that greets the user by name (from command line argument).

### Exercise 2: Temperature Converter

Create a program that converts between Celsius, Fahrenheit, and Kelvin using functions with multiple return values.

### Exercise 3: FizzBuzz with a Twist

Implement FizzBuzz but use:

- A function that returns both the result and a boolean indicating if it was modified
- Switch statements
- Named return values

### Exercise 4: Simple Calculator

Create a calculator that:

- Takes two numbers and an operation from command line
- Returns result and error
- Uses functions as values for operations

## Resources

### Official

- [A Tour of Go](https://go.dev/tour/) - Interactive tutorial
- [Effective Go](https://go.dev/doc/effective_go) - Best practices
- [Go by Example](https://gobyexample.com/) - Code snippets

### For Your Background

- [Go for TypeScript Developers](https://www.typescriptlang.org/docs/handbook/intro.html) - Mental model mapping
- [Go vs Rust](https://hyperpolyglot.org/c/go-rust) - Syntax comparison

## Validation Checklist

- [ ] Go installed and `go version` works
- [ ] Understand `go mod` commands
- [ ] Can declare variables using `:=` and `var`
- [ ] Understand zero values
- [ ] Can write functions with multiple return values
- [ ] Understand package visibility (uppercase = public)
- [ ] Comfortable with Go's `for` loop variations
- [ ] Understand basic pointer usage
- [ ] All exercises completed and running

## Common Pitfalls for Your Background

| From       | Go Difference                                           |
| ---------- | ------------------------------------------------------- |
| TypeScript | No classes, no inheritance, no generics until 1.18      |
| Python     | Explicit error handling, no exceptions for flow control |
| PHP        | Compiled, no dynamic includes, strict typing            |
| Rust       | GC instead of ownership, simpler but less safe          |

## Next Phase

Once comfortable with Go fundamentals, proceed to **Phase 02: Data Structures and Memory** to dive deep into Go's slice/map internals and memory model.
