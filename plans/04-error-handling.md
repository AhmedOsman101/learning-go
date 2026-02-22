# Phase 04: Error Handling and Idioms

**Duration**: 1-2 weeks  
**Prerequisites**: Phase 03 completed  
**Practice Directory**: `phase-04-error-handling/`

## Overview

Go takes a fundamentally different approach to error handling than most languages. Instead of exceptions, Go uses explicit error return values. This phase covers Go's error handling philosophy, creating custom errors, wrapping errors, and the panic/recover mechanism.

## Learning Objectives

- Understand Go's error handling philosophy
- Create and use custom error types
- Master error wrapping and inspection
- Learn when to use panic/recover
- Apply error handling best practices
- Use errors.Is, errors.As, and errors.Join

## Topics to Cover

### 1. Error Basics

```go
// The error interface
type error interface {
    Error() string
}

// Creating errors
import "errors"

err1 := errors.New("something went wrong")
err2 := fmt.Errorf("failed to process %s", filename)

// Returning errors
func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// Handling errors
result, err := Divide(10, 0)
if err != nil {
    fmt.Println("Error:", err)
    return err  // Propagate error
}
fmt.Println("Result:", result)

// The idiomatic pattern
func ProcessFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()

    data, err := io.ReadAll(file)
    if err != nil {
        return fmt.Errorf("failed to read file: %w", err)
    }

    // Process data...
    return nil
}
```

**Key Differences from Exceptions**:

- Errors are values, not special control flow
- No try/catch - explicit handling
- Errors are part of the function signature
- Forces you to think about failure paths

### 2. Custom Error Types

```go
// Custom error type
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

// Usage
func ValidateUser(u *User) error {
    if u.Name == "" {
        return &ValidationError{
            Field:   "name",
            Message: "cannot be empty",
        }
    }
    if u.Age < 0 {
        return &ValidationError{
            Field:   "age",
            Message: "cannot be negative",
        }
    }
    return nil
}

// More complex error type
type HTTPError struct {
    StatusCode int
    Message    string
    Cause      error
}

func (e *HTTPError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("HTTP %d: %s - %v", e.StatusCode, e.Message, e.Cause)
    }
    return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

func (e *HTTPError) Unwrap() error {
    return e.Cause
}

// Sentinel errors
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrForbidden    = errors.New("forbidden")
)

func GetUser(id int) (*User, error) {
    if id <= 0 {
        return nil, ErrNotFound
    }
    // ...
    return &User{}, nil
}
```

### 3. Error Wrapping and Inspection

```go
import "errors"

// Wrapping errors (Go 1.13+)
func ReadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }

    config, err := ParseConfig(data)
    if err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }

    return config, nil
}

// errors.Is - check error chain
if errors.Is(err, ErrNotFound) {
    // Handle not found
}

if errors.Is(err, os.ErrNotExist) {
    // File doesn't exist
}

// errors.As - extract specific error type
var httpErr *HTTPError
if errors.As(err, &httpErr) {
    fmt.Printf("HTTP error: %d\n", httpErr.StatusCode)
}

var validationErr *ValidationError
if errors.As(err, &validationErr) {
    fmt.Printf("Field %s is invalid\n", validationErr.Field)
}

// errors.Join (Go 1.20+) - combine multiple errors
func ValidateAll(user *User) error {
    var errs []error

    if user.Name == "" {
        errs = append(errs, &ValidationError{Field: "name", Message: "required"})
    }
    if user.Email == "" {
        errs = append(errs, &ValidationError{Field: "email", Message: "required"})
    }
    if user.Age < 0 {
        errs = append(errs, &ValidationError{Field: "age", Message: "invalid"})
    }

    return errors.Join(errs...)  // Returns nil if no errors
}

// Checking joined errors
err := ValidateAll(user)
if err != nil {
    // Check individual errors
    var validationErr *ValidationError
    if errors.As(err, &validationErr) {
        fmt.Println(validationErr.Field)
    }
}
```

### 4. Error Handling Patterns

```go
// 1. Immediate return
func DoSomething() error {
    result, err := step1()
    if err != nil {
        return err
    }

    err = step2(result)
    if err != nil {
        return err
    }

    return nil
}

// 2. Wrap with context
func DoSomething() error {
    result, err := step1()
    if err != nil {
        return fmt.Errorf("step1 failed: %w", err)
    }

    if err := step2(result); err != nil {
        return fmt.Errorf("step2 failed: %w", err)
    }

    return nil
}

// 3. Custom error types for different handling
func Process(id int) error {
    user, err := GetUser(id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            return nil  // Not an error, just doesn't exist
        }
        return fmt.Errorf("get user: %w", err)
    }

    if err := ProcessUser(user); err != nil {
        var validationErr *ValidationError
        if errors.As(err, &validationErr) {
            // Log validation errors differently
            log.Printf("validation failed: %s", validationErr.Field)
            return err
        }
        return fmt.Errorf("process user: %w", err)
    }

    return nil
}

// 4. Retry pattern
func WithRetry(fn func() error, maxAttempts int) error {
    var lastErr error
    for i := 0; i < maxAttempts; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        lastErr = err

        // Check if error is retryable
        var netErr net.Error
        if !errors.As(err, &netErr) || !netErr.Timeout() {
            return err  // Not retryable
        }

        time.Sleep(time.Second * time.Duration(i+1))
    }
    return fmt.Errorf("after %d attempts: %w", maxAttempts, lastErr)
}

// 5. Error group for concurrent operations
import "golang.org/x/sync/errgroup"

func ProcessConcurrently(items []Item) error {
    g, ctx := errgroup.WithContext(context.Background())

    for _, item := range items {
        item := item  // Capture loop variable
        g.Go(func() error {
            return processItem(ctx, item)
        })
    }

    return g.Wait()  // Returns first error
}
```

### 5. Panic and Recover

```go
// Panic - for unrecoverable errors
func MustCompile(pattern string) *regexp.Regexp {
    re, err := regexp.Compile(pattern)
    if err != nil {
        panic(err)  // Program crashes if pattern is invalid
    }
    return re
}

// Recover - catch panics
func SafeOperation() (err error) {
    defer func() {
        if r := recover(); r != nil {
            // Convert panic to error
            err = fmt.Errorf("panic recovered: %v", r)
            // Log stack trace
            debug.PrintStack()
        }
    }()

    // Code that might panic
    riskyOperation()
    return nil
}

// HTTP middleware with recovery
func RecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic recovered: %v", err)
                http.Error(w, "Internal Server Error", 500)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// When to use panic:
// 1. During initialization (must fail fast)
// 2. In "Must*" functions (caller guarantees input)
// 3. Truly unrecoverable states
//
// When NOT to use panic:
// 1. Normal error conditions
// 2. Input validation (return error instead)
// 3. Expected failures (file not found, etc.)
```

### 6. Practical Error Types

```go
// Multi-error type
type MultiError struct {
    Errors []error
}

func (e *MultiError) Error() string {
    msgs := make([]string, len(e.Errors))
    for i, err := range e.Errors {
        msgs[i] = err.Error()
    }
    return strings.Join(msgs, "; ")
}

func (e *MultiError) Add(err error) {
    e.Errors = append(e.Errors, err)
}

// Timeout error
type TimeoutError struct {
    Operation string
    Duration  time.Duration
}

func (e *TimeoutError) Error() string {
    return fmt.Sprintf("%s timed out after %v", e.Operation, e.Duration)
}

func (e *TimeoutError) Timeout() bool { return true }
func (e *TimeoutError) Temporary() bool { return true }

// Check for timeout
if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
    // Handle timeout
}

// Error with stack trace
type StackError struct {
    Err   error
    Stack []byte
}

func NewStackError(err error) *StackError {
    return &StackError{
        Err:   err,
        Stack: debug.Stack(),
    }
}

func (e *StackError) Error() string {
    return fmt.Sprintf("%s\nStack:\n%s", e.Err, e.Stack)
}

func (e *StackError) Unwrap() error {
    return e.Err
}
```

### 7. Logging Errors

```go
import (
    "log/slog"
)

// Structured error logging
func ProcessOrder(orderID string) error {
    order, err := GetOrder(orderID)
    if err != nil {
        slog.Error("failed to get order",
            "order_id", orderID,
            "error", err,
        )
        return fmt.Errorf("get order %s: %w", orderID, err)
    }

    if err := ValidateOrder(order); err != nil {
        slog.Warn("order validation failed",
            "order_id", orderID,
            "error", err,
        )
        return fmt.Errorf("validate order: %w", err)
    }

    slog.Info("order processed",
        "order_id", orderID,
        "total", order.Total,
    )
    return nil
}

// Error levels
// - Error: Something went wrong, needs attention
// - Warn: Unexpected but handled, might indicate problem
// - Info: Normal operations
// - Debug: Detailed info for debugging
```

## Hands-on Exercises

Create the following programs in `phase-04-error-handling/`:

### Exercise 1: Custom Error Hierarchy

Create a hierarchy of errors for a user management system:

- NotFoundError
- ValidationError
- DuplicateError
- DatabaseError

### Exercise 2: Error Wrapping Chain

Build a multi-layer application (handler -> service -> repository) with proper error wrapping at each layer.

### Exercise 3: Retry Mechanism

Implement a retry mechanism with:

- Exponential backoff
- Max retries
- Retryable error detection
- Context cancellation support

### Exercise 4: Error Reporting

Create an error reporter that:

- Collects errors with stack traces
- Groups similar errors
- Outputs JSON format for external systems

## Resources

### Official

- [Error Handling in Go](https://go.dev/blog/error-handling)
- [Go 1.13 Errors](https://go.dev/blog/go1.13-errors)
- [Package errors](https://pkg.go.dev/errors)

### Best Practices

- [Error Handling in Upspin](https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html)
- [Don't just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)

## Validation Checklist

- [ ] Understand error interface
- [ ] Can create custom error types
- [ ] Know how to wrap errors with %w
- [ ] Can use errors.Is and errors.As
- [ ] Understand when to use panic/recover
- [ ] Know error handling best practices
- [ ] All exercises completed

## Common Pitfalls

| Mistake                 | Consequence      | Fix                  |
| ----------------------- | ---------------- | -------------------- |
| Ignoring errors         | Silent failures  | Always handle errors |
| Using %v instead of %w  | Lost error chain | Use %w for wrapping  |
| Panic for normal errors | Crashes          | Return error instead |
| Not adding context      | Hard to debug    | Wrap with context    |

## Next Phase

Proceed to **Phase 05: Concurrency Fundamentals** to learn Go's powerful concurrency primitives.
