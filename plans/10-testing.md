# Phase 10: Testing and Benchmarking

**Duration**: 2 weeks  
**Prerequisites**: Phase 09 completed  
**Practice Directory**: `phase-10-testing/`

## Overview

Testing is a first-class citizen in Go. The built-in testing package, combined with tools like testify and mockery, provides a robust testing ecosystem. This phase covers unit testing, integration testing, benchmarks, and fuzzing.

## Learning Objectives

- Write effective unit tests
- Use table-driven tests
- Implement mocks and stubs
- Write integration tests
- Create benchmarks
- Use fuzzing for finding bugs
- Measure and improve code coverage

## Topics to Cover

### 1. Basic Testing

```go
// math.go
package math

func Add(a, b int) int {
    return a + b
}

func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// math_test.go
package math

import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }
}

func TestDivide(t *testing.T) {
    t.Run("valid division", func(t *testing.T) {
        result, err := Divide(10, 2)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if result != 5 {
            t.Errorf("Divide(10, 2) = %f; want 5", result)
        }
    })

    t.Run("division by zero", func(t *testing.T) {
        _, err := Divide(10, 0)
        if err == nil {
            t.Error("expected error for division by zero")
        }
    })
}

// Run tests
// go test ./...
// go test -v ./...
// go test -run TestDivide ./...
// go test -run TestDivide/valid ./...
```

### 2. Table-Driven Tests

```go
func TestAddTable(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -2, -3, -5},
        {"mixed numbers", -2, 3, 1},
        {"zeros", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d",
                    tt.a, tt.b, result, tt.expected)
            }
        })
    }
}

// Complex table test
func TestValidateUser(t *testing.T) {
    tests := []struct {
        name    string
        user    User
        wantErr bool
        errType error
    }{
        {
            name:    "valid user",
            user:    User{Name: "John", Email: "john@example.com", Age: 25},
            wantErr: false,
        },
        {
            name:    "empty name",
            user:    User{Name: "", Email: "john@example.com", Age: 25},
            wantErr: true,
            errType: ErrInvalidName,
        },
        {
            name:    "invalid email",
            user:    User{Name: "John", Email: "invalid", Age: 25},
            wantErr: true,
            errType: ErrInvalidEmail,
        },
        {
            name:    "negative age",
            user:    User{Name: "John", Email: "john@example.com", Age: -1},
            wantErr: true,
            errType: ErrInvalidAge,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateUser(tt.user)

            if tt.wantErr {
                if err == nil {
                    t.Fatal("expected error, got nil")
                }
                if tt.errType != nil && !errors.Is(err, tt.errType) {
                    t.Errorf("expected error %v, got %v", tt.errType, err)
                }
            } else {
                if err != nil {
                    t.Errorf("unexpected error: %v", err)
                }
            }
        })
    }
}
```

### 3. Testify

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/suite"
)

func TestAddTestify(t *testing.T) {
    // assert - continues on failure
    assert.Equal(t, 5, Add(2, 3), "should add correctly")
    assert.NotEqual(t, 6, Add(2, 3))

    // require - stops on failure
    result, err := Divide(10, 2)
    require.NoError(t, err)
    assert.Equal(t, 5.0, result)

    // More assertions
    assert.True(t, true)
    assert.False(t, false)
    assert.Nil(t, nil)
    assert.NotNil(t, &User{})
    assert.Len(t, []int{1, 2, 3}, 3)
    assert.Contains(t, "hello world", "world")
    assert.InDelta(t, 1.0, 1.1, 0.2)  // Float comparison with tolerance

    // Error assertions
    _, err = Divide(10, 0)
    assert.Error(t, err)
    assert.EqualError(t, err, "division by zero")
    assert.ErrorIs(t, err, ErrDivisionByZero)
}

// Test suite
type UserTestSuite struct {
    suite.Suite
    db   *sql.DB
    repo *UserRepository
}

func (s *UserTestSuite) SetupSuite() {
    // Run once before all tests
    s.db = setupTestDB()
    s.repo = NewUserRepository(s.db)
}

func (s *UserTestSuite) TearDownSuite() {
    // Run once after all tests
    s.db.Close()
}

func (s *UserTestSuite) SetupTest() {
    // Run before each test
    cleanupDB(s.db)
}

func (s *UserTestSuite) TestCreateUser() {
    user := &User{Name: "John", Email: "john@example.com"}

    err := s.repo.Create(user)
    s.NoError(err)
    s.NotZero(user.ID)
}

func (s *UserTestSuite) TestGetUser() {
    // Create test user
    created := &User{Name: "Jane", Email: "jane@example.com"}
    s.repo.Create(created)

    // Get user
    user, err := s.repo.Get(created.ID)
    s.NoError(err)
    s.Equal("Jane", user.Name)
}

func TestUserTestSuite(t *testing.T) {
    suite.Run(t, new(UserTestSuite))
}
```

### 4. Mocking

```go
// Interface to mock
type UserRepository interface {
    Get(id int64) (*User, error)
    Create(user *User) error
    Update(user *User) error
    Delete(id int64) error
}

// Manual mock
type MockUserRepository struct {
    users map[int64]*User
    err   error
}

func (m *MockUserRepository) Get(id int64) (*User, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.users[id], nil
}

// Using testify/mock
import "github.com/stretchr/testify/mock"

type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Get(id int64) (*User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) Create(user *User) error {
    args := m.Called(user)
    return args.Error(0)
}

// Using the mock
func TestUserService_Get(t *testing.T) {
    mockRepo := new(MockRepository)
    service := NewUserService(mockRepo)

    // Setup expectations
    expectedUser := &User{ID: 1, Name: "John"}
    mockRepo.On("Get", int64(1)).Return(expectedUser, nil)

    // Test
    user, err := service.Get(1)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedUser, user)
    mockRepo.AssertExpectations(t)
}

// Using mockery to generate mocks
//go:generate mockery --name=UserRepository --output=mocks
// Run: go generate ./...
```

### 5. HTTP Testing

```go
import (
    "net/http"
    "net/http/httptest"
)

func TestHandler(t *testing.T) {
    handler := &UserHandler{service: mockService}

    req := httptest.NewRequest("GET", "/users/1", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)

    var response User
    json.Unmarshal(rec.Body.Bytes(), &response)
    assert.Equal(t, "John", response.Name)
}

func TestCreateUser(t *testing.T) {
    handler := &UserHandler{service: mockService}

    body := `{"name": "John", "email": "john@example.com"}`
    req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusCreated, rec.Code)
}

// Testing with httptest.Server
func TestHTTPClient(t *testing.T) {
    // Create test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/api/users", r.URL.Path)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`[{"id": 1, "name": "John"}]`))
    }))
    defer server.Close()

    // Use server.URL as base URL
    client := NewClient(server.URL)
    users, err := client.GetUsers()

    assert.NoError(t, err)
    assert.Len(t, users, 1)
}
```

### 6. Integration Testing

```go
// Using testcontainers
import "github.com/testcontainers/testcontainers-go"

func TestWithPostgres(t *testing.T) {
    ctx := context.Background()

    // Start PostgreSQL container
    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "postgres:15",
            ExposedPorts: []string{"5432/tcp"},
            Env: map[string]string{
                "POSTGRES_USER":     "test",
                "POSTGRES_PASSWORD": "test",
                "POSTGRES_DB":       "test",
            },
            WaitingFor: wait.ForLog("database system is ready to accept connections"),
        },
        Started: true,
    })
    require.NoError(t, err)
    defer container.Terminate(ctx)

    // Get connection details
    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")

    dsn := fmt.Sprintf("postgres://test:test@%s:%s/test?sslmode=disable", host, port.Port())
    db, _ := sql.Open("postgres", dsn)
    defer db.Close()

    // Run migrations
    runMigrations(db)

    // Test
    repo := NewUserRepository(db)
    user := &User{Name: "John", Email: "john@example.com"}
    err = repo.Create(user)
    assert.NoError(t, err)
}

// Build tags for integration tests
// +build integration

func TestIntegration(t *testing.T) {
    // This test only runs with: go test -tags=integration ./...
}

// Skip short tests
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    // ...
}

// Run with: go test -short ./...  (skips integration tests)
```

### 7. Benchmarks

```go
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}

func BenchmarkFibonacci(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Fibonacci(20)
    }
}

// Benchmark with setup
func BenchmarkProcessUser(b *testing.B) {
    user := &User{Name: "John", Email: "john@example.com"}

    b.ResetTimer()  // Reset timer after setup
    for i := 0; i < b.N; i++ {
        ProcessUser(user)
    }
}

// Parallel benchmark
func BenchmarkProcessParallel(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            ProcessUser(&User{Name: "John"})
        }
    })
}

// Sub-benchmarks
func BenchmarkJSON(b *testing.B) {
    data := User{Name: "John", Email: "john@example.com"}

    b.Run("marshal", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            json.Marshal(data)
        }
    })

    b.Run("unmarshal", func(b *testing.B) {
        bytes, _ := json.Marshal(data)
        var u User
        for i := 0; i < b.N; i++ {
            json.Unmarshal(bytes, &u)
        }
    })
}

// Run benchmarks
// go test -bench=. ./...
// go test -bench=BenchmarkAdd ./...
// go test -bench=. -benchmem ./...  (show memory allocations)
// go test -bench=. -benchtime=5s ./...
```

### 8. Fuzzing (Go 1.18+)

```go
func FuzzAdd(f *testing.F) {
    // Add seed corpus
    f.Add(1, 2)
    f.Add(-1, 1)
    f.Add(0, 0)

    f.Fuzz(func(t *testing.T, a, b int) {
        result := Add(a, b)

        // Property: addition is commutative
        if Add(b, a) != result {
            t.Errorf("Add is not commutative: Add(%d, %d) != Add(%d, %d)", a, b, b, a)
        }

        // Property: result is always >= min(a, b) when both positive
        if a >= 0 && b >= 0 && result < a {
            t.Errorf("unexpected result: Add(%d, %d) = %d", a, b, result)
        }
    })
}

func FuzzReverse(f *testing.F) {
    f.Add("hello")
    f.Add("")
    f.Add("a")

    f.Fuzz(func(t *testing.T, s string) {
        reversed := Reverse(s)
        doubleReversed := Reverse(reversed)

        // Property: reversing twice returns original
        if doubleReversed != s {
            t.Errorf("Reverse(Reverse(%q)) = %q, want %q", s, doubleReversed, s)
        }
    })
}

// Run fuzzing
// go test -fuzz=FuzzAdd ./...
// go test -fuzz=FuzzReverse -fuzztime=30s ./...

// Fuzz findings are saved to testdata/fuzz/
```

### 9. Coverage

```go
// Run with coverage
// go test -cover ./...
// go test -coverprofile=coverage.out ./...
// go tool cover -func=coverage.out
// go tool cover -html=coverage.out -o coverage.html

// Coverage in CI
func TestMain(m *testing.M) {
    // Setup
    os.Exit(m.Run())
    // Teardown
}
```

## Hands-on Exercises

Create the following programs in `phase-10-testing/`:

### Exercise 1: Unit Test Suite

Write comprehensive tests for a calculator package:

- Table-driven tests
- Edge cases
- Error conditions
- 100% coverage

### Exercise 2: Mock-Based Testing

Create a service layer with:

- Repository interface
- Generated mocks
- Service tests using mocks
- Verify all interactions

### Exercise 3: Integration Tests

Build integration tests for a user API:

- Test database setup
- HTTP endpoint testing
- Transaction rollback after tests
- CI-ready configuration

### Exercise 4: Benchmark Suite

Create benchmarks for:

- String operations
- JSON parsing
- Concurrent operations
- Compare implementations

## Resources

### Official

- [Testing package](https://pkg.go.dev/testing)
- [Go Testing Blog](https://go.dev/blog/testing)

### Libraries

- [Testify](https://github.com/stretchr/testify)
- [Mockery](https://github.com/vektra/mockery)
- [Testcontainers](https://github.com/testcontainers/testcontainers-go)

## Validation Checklist

- [ ] Can write basic tests
- [ ] Can use table-driven tests
- [ ] Can use testify assertions
- [ ] Can create and use mocks
- [ ] Can test HTTP handlers
- [ ] Can write integration tests
- [ ] Can write benchmarks
- [ ] Can use fuzzing
- [ ] All exercises completed

## Next Phase

Proceed to **Phase 11: Tooling and Ecosystem** to master Go's development tools.
