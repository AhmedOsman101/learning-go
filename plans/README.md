# Golang Learning Path: From Basics to Mastery

A comprehensive learning plan for experienced developers transitioning to Go.

## Overview

This learning path is designed for developers who already know TypeScript, PHP, Python, Bash, and Rust basics. It skips fundamental programming concepts and focuses on Go-specific idioms, patterns, and best practices.

## Learning Phases

| Phase | Title                                                        | Duration  | Focus                                   |
| ----- | ------------------------------------------------------------ | --------- | --------------------------------------- |
| 01    | [Go Fundamentals](./01-go-fundamentals.md)                   | 1-2 weeks | Syntax, tooling, project structure      |
| 02    | [Data Structures & Memory](./02-data-structures.md)          | 1-2 weeks | Slices, maps, structs, memory model     |
| 03    | [Interfaces & Type System](./03-interfaces.md)               | 1-2 weeks | Interfaces, generics, type assertions   |
| 04    | [Error Handling](./04-error-handling.md)                     | 1-2 weeks | Error patterns, wrapping, panic/recover |
| 05    | [Concurrency Fundamentals](./05-concurrency-fundamentals.md) | 2-3 weeks | Goroutines, channels, sync primitives   |
| 06    | [Advanced Concurrency](./06-advanced-concurrency.md)         | 2-3 weeks | Context, patterns, debugging            |
| 07    | [Standard Library](./07-stdlib.md)                           | 2-3 weeks | io, net/http, encoding, time            |
| 08    | [Web Services](./08-web-services.md)                         | 2-3 weeks | REST APIs, middleware, auth, gRPC       |
| 09    | [Databases](./09-databases.md)                               | 2-3 weeks | SQL, ORM, migrations, Redis             |
| 10    | [Testing & Benchmarking](./10-testing.md)                    | 2 weeks   | Unit tests, mocks, benchmarks, fuzzing  |
| 11    | [Tooling & Ecosystem](./11-tooling.md)                       | 1-2 weeks | Modules, linting, profiling, CI/CD      |
| 12    | [Advanced Topics](./12-advanced.md)                          | 2-3 weeks | Reflection, unsafe, cgo, optimization   |

**Total Estimated Duration**: 20-30 weeks

## Directory Structure

```
learning/
├── plans/                          # Learning plans (this directory)
│   ├── 01-go-fundamentals.md
│   ├── 02-data-structures.md
│   └── ...
├── phase-01-fundamentals/          # Practice exercises for phase 1
├── phase-02-data-structures/       # Practice exercises for phase 2
├── phase-03-interfaces/
├── phase-04-error-handling/
├── phase-05-concurrency-basics/
├── phase-06-concurrency-patterns/
├── phase-07-stdlib/
├── phase-08-web-services/
├── phase-09-databases/
├── phase-10-testing/
├── phase-11-tooling/
└── phase-12-advanced/
```

## How to Use This Plan

1. **Read the phase document** in `plans/xx-title.md`
2. **Complete the exercises** in the corresponding `phase-xx-title/` directory
3. **Check off items** in the validation checklist
4. **Proceed to the next phase** when ready

## Key Differences from Other Languages

| From       | Key Go Differences                                 |
| ---------- | -------------------------------------------------- |
| TypeScript | No classes, no inheritance, implicit interfaces    |
| Python     | Explicit error handling, compiled, strict typing   |
| PHP        | No dynamic includes, compiled binaries, goroutines |
| Rust       | GC instead of ownership, simpler but less safe     |
| Bash       | Compiled, cross-platform, rich standard library    |

## Recommended Resources

### Official

- [A Tour of Go](https://go.dev/tour/) - Interactive tutorial
- [Effective Go](https://go.dev/doc/effective_go) - Best practices
- [Go by Example](https://gobyexample.com/) - Code snippets

### Books

- [The Go Programming Language](https://www.gopl.io/)
- [Concurrency in Go](https://www.oreilly.com/library/view/concurrency-in-go/9781491941294/)
- [100 Go Mistakes](https://100go.co/)

### Community

- [Go Forum](https://forum.golangbridge.org/)
- [Gophers Slack](https://gophers.slack.com/)
- [r/golang](https://reddit.com/r/golang)

## Tips for Success

1. **Write code daily** - Even small programs help reinforce concepts
2. **Read the standard library** - It's well-written and idiomatic
3. **Use the race detector** - `go run -race` catches concurrency bugs
4. **Format your code** - `go fmt` is non-negotiable
5. **Embrace simplicity** - Go favors simple, readable code

---

Good luck on your Go journey! 🎉
