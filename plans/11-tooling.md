# Phase 11: Tooling and Ecosystem

**Duration**: 1-2 weeks  
**Prerequisites**: Phase 10 completed  
**Practice Directory**: `phase-11-tooling/`

## Overview

Go has excellent built-in tooling and a rich ecosystem of third-party tools. This phase covers module management, code generation, linting, profiling, documentation, and CI/CD integration.

## Learning Objectives

- Master Go modules and dependency management
- Use code generation tools effectively
- Configure and use linters
- Profile and optimize Go programs
- Generate documentation
- Set up CI/CD pipelines

## Topics to Cover

### 1. Go Modules

```bash
# Initialize module
go mod init github.com/user/project

# Add dependency
go get github.com/gin-gonic/gin
go get github.com/gin-gonic/gin@v1.9.0  # Specific version
go get github.com/gin-gonic/gin@latest

# Update dependencies
go get -u ./...           # Update all
go get -u=patch ./...     # Update patches only
go mod tidy               # Clean up go.mod and go.sum

# View dependencies
go list -m all
go list -m -versions github.com/gin-gonic/gin

# Verify dependencies
go mod verify

# Vendor mode
go mod vendor

# Replace directive (for local development)
# go.mod
module github.com/user/project

go 1.21

require github.com/user/mylib v1.0.0

replace github.com/user/mylib => ../mylib

# Exclude specific versions
exclude github.com/user/badlib v1.0.0

# Workspace (Go 1.21+)
go work init
go work use ./module1
go work use ./module2
```

### 2. Code Generation

```go
//go:generate directives

// Generate mocks
//go:generate mockery --name=Repository --output=mocks --outpkg=mocks

// Generate stringer
//go:generate go run golang.org/x/tools/cmd/stringer -type=Status

type Status int

const (
    StatusUnknown Status = iota
    StatusActive
    StatusInactive
)

// Generate code from templates
//go:generate go run github.com/abice/go-enum -f=$GOFILE

// Generate protobuf
//go:generate protoc --go_out=. --go_opt=paths=source_relative proto/*.proto

// Generate SQL code (sqlc)
//go:generate sqlc generate

// Run generation
// go generate ./...
// go generate ./path/to/package
```

### 3. Linting

```bash
# Built-in vet
go vet ./...

# Staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...

# golangci-lint (aggregates multiple linters)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run

# .golangci.yml
linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - ineffassign
    - typecheck
    - gosimple
    - goconst
    - gocyclo
    - dupl
    - misspell

linters-settings:
  gocyclo:
    min-complexity: 15
  goconst:
    min-len: 3
    min-occurrences: 3

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - dupl
        - gosec

run:
  timeout: 5m
  skip-dirs:
    - vendor
    - mocks
```

### 4. Formatting

```bash
# Built-in formatter
go fmt ./...
gofmt -w .

# goimports (also manages imports)
go install golang.org/x/tools/cmd/goimports@latest
goimports -w .

# gofumpt (stricter formatting)
go install mvdan.cc/gofumpt@latest
gofumpt -w .

# Editor integration
# VS Code: install Go extension
# GoLand: built-in support
```

### 5. Profiling

```go
import (
    "runtime/pprof"
    "net/http"
    _ "net/http/pprof"
)

// CPU profiling
func main() {
    f, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // Your code...
}

// Memory profiling
func main() {
    // Your code...

    f, _ := os.Create("mem.prof")
    pprof.WriteHeapProfile(f)
    f.Close()
}

// HTTP pprof endpoint
func main() {
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()

    // Your code...
}

// Access profiles:
// http://localhost:6060/debug/pprof/
// http://localhost:6060/debug/pprof/profile?seconds=30
// http://localhost:6060/debug/pprof/heap
```

```bash
# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof

# Interactive commands
# (pprof) top
# (pprof) top -cum
# (pprof) list functionName
# (pprof) web  # Requires graphviz

# From URL
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
go tool pprof http://localhost:6060/debug/pprof/heap

# Flame graphs
go tool pprof -http=:8080 cpu.prof
```

### 6. Tracing

```go
import (
    "runtime/trace"
)

func main() {
    f, _ := os.Create("trace.out")
    trace.Start(f)
    defer trace.Stop()

    // Your code...
}

// Custom trace regions
func process() {
    ctx, task := trace.NewTask(context.Background(), "process")
    defer task.End()

    trace.WithRegion(ctx, "step1", step1)
    trace.WithRegion(ctx, "step2", step2)

    trace.Log(ctx, "key", "value")
}
```

```bash
# View trace
go tool trace trace.out

# Opens browser with:
# - View trace timeline
# - Goroutine analysis
# - Network blocking profile
# - Synchronization blocking profile
# - Syscall blocking profile
# - Scheduler latency profile
```

### 7. Documentation

```go
// Package comment (first sentence appears in package list)
// Package user provides user management functionality.
// It supports CRUD operations and authentication.
package user

// User represents a system user.
// Users have a unique ID, name, and email address.
type User struct {
    // ID is the unique identifier for the user.
    ID int64

    // Name is the user's display name.
    // It must be between 1 and 100 characters.
    Name string

    // Email is the user's email address.
    // It must be a valid email format.
    Email string
}

// Create creates a new user with the given name and email.
// It returns an error if the email is already in use.
//
// Example:
//
//  user, err := Create("John", "john@example.com")
//  if err != nil {
//      log.Fatal(err)
//  }
func Create(name, email string) (*User, error) {
    // ...
}

// Constants and variables
const (
    // MaxNameLength is the maximum allowed name length.
    MaxNameLength = 100
)
```

```bash
# View documentation
go doc
go doc user
go doc user.User
go doc user.Create

# Generate HTML documentation
go install golang.org/x/tools/cmd/godoc@latest
godoc -http=:6060

# Open http://localhost:6060/pkg/

# Gold (alternative)
go install go.abhg.dev/gold@latest
gold -http=:6060
```

### 8. Build and Release

```bash
# Build
go build -o myapp ./cmd/server
go build -ldflags="-s -w" ./cmd/server  # Strip debug info

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o myapp-linux ./cmd/server
GOOS=darwin GOARCH=arm64 go build -o myapp-darwin ./cmd/server
GOOS=windows GOARCH=amd64 go build -o myapp.exe ./cmd/server

# Version injection
# main.go
var (
    version   = "dev"
    commit    = "none"
    buildTime = "unknown"
)

# Build
go build -ldflags="-X main.version=1.0.0 -X main.commit=$(git rev-parse HEAD) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" ./cmd/server

# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]

# Goreleaser
# .goreleaser.yml
project_name: myapp
builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}
archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
```

### 9. CI/CD

```yaml
# GitHub Actions
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

  build:
    runs-on: ubuntu-latest
    needs: [test, lint]
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: Build
        run: go build -v ./...

      - name: Build Docker image
        run: docker build -t myapp:${{ github.sha }} .
```

### 10. Development Tools

```bash
# Air - Live reload
go install github.com/cosmtrek/air@latest
air

# .air.toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ./cmd/server"
  bin = "./tmp/main"
  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_dir = ["assets", "tmp", "vendor"]

# Delve - Debugger
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug ./cmd/server
dlv test ./...

# Debugger commands
# (dlv) break main.main
# (dlv) continue
# (dlv) next
# (dlv) step
# (dlv) print variable
# (dlv) goroutines
# (dlv) stack

# Go repl
go install github.com/x-motemen/ghq@latest

# Stringer (generate String methods)
go install golang.org/x/tools/cmd/stringer@latest

//go:generate stringer -type=Status

# Mockery
go install github.com/vektra/mockery/v2@latest

# SQLC
go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

# Wire (dependency injection)
go install github.com/google/wire/cmd/wire@latest
```

## Hands-on Exercises

Create the following programs in `phase-11-tooling/`:

### Exercise 1: Module Management

Set up a multi-module workspace:

- Main application module
- Shared library module
- Proper versioning
- Replace directives for local development

### Exercise 2: CI Pipeline

Create a complete CI pipeline:

- Test matrix (multiple Go versions)
- Linting with golangci-lint
- Coverage reporting
- Docker builds

### Exercise 3: Profiling Exercise

Profile and optimize a program:

- Identify bottlenecks
- Memory leaks
- CPU hotspots
- Document improvements

### Exercise 4: Code Generation

Set up code generation:

- Mock generation
- String method generation
- SQL code generation
- Wire dependency injection

## Resources

### Official

- [Go Modules](https://go.dev/blog/using-go-modules)
- [Go Blog: Profiling](https://go.dev/blog/pprof)
- [Go Documentation](https://go.dev/doc/)

### Tools

- [golangci-lint](https://golangci-lint.run/)
- [Air](https://github.com/cosmtrek/air)
- [Delve](https://github.com/go-delve/delve)
- [Goreleaser](https://goreleaser.com/)

## Validation Checklist

- [ ] Can manage dependencies with go mod
- [ ] Can use go:generate directives
- [ ] Can configure golangci-lint
- [ ] Can profile CPU and memory
- [ ] Can write documentation
- [ ] Can set up CI/CD
- [ ] All exercises completed

## Next Phase

Proceed to **Phase 12: Advanced Topics** to explore reflection, unsafe, cgo, and performance optimization.
