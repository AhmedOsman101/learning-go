# Phase 07: Standard Library Deep Dive

**Duration**: 2-3 weeks  
**Prerequisites**: Phase 06 completed  
**Practice Directory**: `phase-07-stdlib/`

## Overview

Go's standard library is one of its greatest strengths. Unlike many languages where you immediately reach for third-party packages, Go's stdlib covers most common needs. This phase explores the most important packages in depth.

## Learning Objectives

- Master the io package and its interfaces
- Work with files and directories effectively
- Parse and generate JSON, XML, and other formats
- Use the net/http package for clients and servers
- Understand the encoding packages
- Work with time, strings, and regular expressions

## Topics to Cover

### 1. IO Package

```go
import "io"

// Core interfaces
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}

type Seeker interface {
    Seek(offset int64, whence int) (int64, error)
}

// Composed interfaces
type ReadWriter interface { Reader; Writer }
type ReadCloser interface { Reader; Closer }
type WriteCloser interface { Writer; Closer }
type ReadWriteCloser interface { Reader; Writer; Closer }

// Common operations
io.Copy(dst, src)           // Copy from src to dst
io.CopyN(dst, src, n)       // Copy n bytes
io.CopyBuffer(dst, src, buf) // Copy with custom buffer

io.ReadAll(src)             // Read all to []byte
io.LimitReader(r, n)        // Reader that stops after n bytes
io.MultiReader(r1, r2)      // Read from multiple readers sequentially
io.MultiWriter(w1, w2)      // Write to multiple writers
io.TeeReader(r, w)          // Read from r, also write to w
io.Pipe()                   // In-memory pipe (r, w)

// String readers/writers
r := strings.NewReader("hello")
w := &strings.Builder{}
io.Copy(w, r)
fmt.Println(w.String())

// Discard (like /dev/null)
io.Copy(io.Discard, reader)

// NopCloser - wrap Reader as ReadCloser
rc := io.NopCloser(reader)
```

### 2. File Operations

```go
import "os"

// Opening files
file, err := os.Open("file.txt")           // Read-only
file, err := os.OpenFile("file.txt", os.O_RDWR|os.O_CREATE, 0644)

// Reading
data, err := os.ReadFile("file.txt")       // Read entire file
n, err := file.Read(buf)                   // Read into buffer

// Writing
err := os.WriteFile("file.txt", data, 0644)
n, err := file.Write(data)
n, err := file.WriteString("hello")

// File info
info, err := os.Stat("file.txt")
fmt.Println(info.Name(), info.Size(), info.ModTime(), info.Mode())

// Check if exists
if _, err := os.Stat("file.txt"); os.IsNotExist(err) {
    // File doesn't exist
}

// Create directory
os.Mkdir("dir", 0755)
os.MkdirAll("dir/subdir/deep", 0755)

// Remove
os.Remove("file.txt")
os.RemoveAll("dir")

// Working with paths
import "path/filepath"

filepath.Join("dir", "subdir", "file.txt")  // "dir/subdir/file.txt"
filepath.Base("/path/to/file.txt")          // "file.txt"
filepath.Dir("/path/to/file.txt")           // "/path/to"
filepath.Ext("file.txt")                    // ".txt"
filepath.Abs("file.txt")                    // Full path
filepath.Walk("dir", func(path string, info os.FileInfo, err error) error {
    fmt.Println(path)
    return nil
})

// Temp files
file, err := os.CreateTemp("", "prefix")    // Creates temp file
dir, err := os.MkdirTemp("", "prefix")      // Creates temp dir
defer os.Remove(file.Name())

// File modes
file.Chmod(0644)
file.Chown(uid, gid)
file.Sync()  // Flush to disk
```

### 3. Encoding/Decoding

```go
import "encoding/json"

// JSON encoding
type Person struct {
    Name  string `json:"name"`
    Age   int    `json:"age,omitempty"`
    Email string `json:"email,omitempty"`
}

// Struct to JSON
p := Person{Name: "John", Age: 30}
data, err := json.Marshal(p)
data, err := json.MarshalIndent(p, "", "  ")  // Pretty print

// JSON to struct
var p2 Person
err := json.Unmarshal(data, &p2)

// Streaming
encoder := json.NewEncoder(file)
encoder.Encode(p)

decoder := json.NewDecoder(file)
decoder.Decode(&p)

// Unknown structure
var result map[string]interface{}
json.Unmarshal(data, &result)

// Raw JSON
var raw json.RawMessage
json.Unmarshal(data, &raw)
// raw contains the raw JSON bytes

// Custom marshaling
func (p *Person) MarshalJSON() ([]byte, error) {
    type Alias Person
    return json.Marshal(&struct {
        *Alias
        UpperName string `json:"upperName"`
    }{
        Alias:     (*Alias)(p),
        UpperName: strings.ToUpper(p.Name),
    })
}

// XML
import "encoding/xml"

type PersonXML struct {
    XMLName xml.Name `xml:"person"`
    Name    string   `xml:"name"`
    Age     int      `xml:"age"`
}

data, _ := xml.Marshal(p)
xml.Unmarshal(data, &p)

// Base64
import "encoding/base64"

encoded := base64.StdEncoding.EncodeToString(data)
decoded, _ := base64.StdEncoding.DecodeString(encoded)

// Hex
import "encoding/hex"

encoded := hex.EncodeToString(data)
decoded, _ := hex.DecodeString(encoded)

// Binary
import "encoding/binary"

var buf [4]byte
binary.LittleEndian.PutUint32(buf[:], 12345)
val := binary.LittleEndian.Uint32(buf[:])

// Gob (Go-specific binary encoding)
import "encoding/gob"

encoder := gob.NewEncoder(file)
encoder.Encode(p)

decoder := gob.NewDecoder(file)
decoder.Decode(&p)
```

### 4. HTTP Package

```go
import "net/http"

// Simple HTTP client
resp, err := http.Get("https://example.com")
resp, err := http.Post("https://example.com", "application/json", body)
resp, err := http.PostForm("https://example.com", url.Values{"key": {"value"}})

// Always close response body
defer resp.Body.Close()
body, _ := io.ReadAll(resp.Body)

// Advanced client
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}

// Request with context
req, _ := http.NewRequestWithContext(ctx, "GET", "https://example.com", nil)
req.Header.Set("Authorization", "Bearer token")
resp, _ := client.Do(req)

// HTTP server
http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "hello"})
})

http.ListenAndServe(":8080", nil)

// Server with options
server := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  60 * time.Second,
}

// Middleware
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL, time.Since(start))
    })
}

// Using ServeMux (Go 1.22+)
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", getUser)        // Method + path pattern
mux.HandleFunc("POST /users", createUser)
mux.HandleFunc("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

// Path values
func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")  // Go 1.22+
    // ...
}

// Request parsing
func handler(w http.ResponseWriter, r *http.Request) {
    // Query params
    query := r.URL.Query()
    page := query.Get("page")

    // Headers
    auth := r.Header.Get("Authorization")

    // Body
    var data Request
    json.NewDecoder(r.Body).Decode(&data)

    // Form data
    r.ParseForm()
    value := r.FormValue("key")

    // Cookies
    cookie, err := r.Cookie("session")
    http.SetCookie(w, &http.Cookie{
        Name:  "session",
        Value: "token",
    })
}

// Response helpers
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, status int, message string) {
    jsonResponse(w, status, map[string]string{"error": message})
}
```

### 5. Time Package

```go
import "time"

// Current time
now := time.Now()
fmt.Println(now.Year(), now.Month(), now.Day())
fmt.Println(now.Hour(), now.Minute(), now.Second())

// Parsing
t, _ := time.Parse("2006-01-02", "2024-01-15")
t, _ := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")

// Formatting (use reference time: Mon Jan 2 15:04:05 MST 2006)
formatted := t.Format("2006-01-02 15:04:05")
fmt.Println(t.Format(time.RFC3339))

// Duration
d := 5 * time.Second
d = 2 * time.Hour
d = 500 * time.Millisecond

// Time arithmetic
future := now.Add(24 * time.Hour)
past := now.Add(-24 * time.Hour)
diff := future.Sub(now)

// Comparing times
if now.Before(future) { }
if now.After(past) { }
if now.Equal(other) { }

// Sleep
time.Sleep(1 * time.Second)

// Ticker
ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()
for t := range ticker.C {
    fmt.Println(t)
}

// Timer
timer := time.NewTimer(5 * time.Second)
<-timer.C  // Wait for timer

// After (one-shot)
select {
case <-time.After(5 * time.Second):
    fmt.Println("timeout")
}

// Since/Until
elapsed := time.Since(start)
remaining := time.Until(deadline)

// Truncate and Round
t = t.Truncate(time.Hour)  // Round down to hour
t = t.Round(time.Minute)   // Round to nearest minute

// Unix timestamp
unix := now.Unix()         // Seconds
unixMilli := now.UnixMilli() // Milliseconds
fromUnix := time.Unix(unix, 0)

// Time zones
loc, _ := time.LoadLocation("America/New_York")
t = now.In(loc)
t = time.Date(2024, 1, 15, 10, 30, 0, 0, loc)

// Local vs UTC
local := now.Local()
utc := now.UTC()
```

### 6. Strings and Text

```go
import "strings"

// Basic operations
strings.Contains("hello", "ell")      // true
strings.HasPrefix("hello", "hel")     // true
strings.HasSuffix("hello", "llo")     // true
strings.Index("hello", "l")           // 2
strings.LastIndex("hello", "l")       // 3
strings.Count("hello", "l")           // 2
strings.Repeat("ab", 3)               // "ababab"
strings.Replace("hello", "l", "L", 1) // "heLlo"
strings.ReplaceAll("hello", "l", "L") // "heLLo"

// Splitting and joining
strings.Split("a,b,c", ",")           // ["a", "b", "c"]
strings.SplitN("a,b,c", ",", 2)       // ["a", "b,c"]
strings.Join([]string{"a", "b"}, ",") // "a,b"

// Trimming
strings.TrimSpace("  hello  ")        // "hello"
strings.Trim("xxhelloxx", "x")        // "hello"
strings.TrimLeft("xxhello", "x")      // "hello"
strings.TrimRight("helloxx", "x")     // "hello"

// Case conversion
strings.ToLower("HELLO")              // "hello"
strings.ToUpper("hello")              // "HELLO"
strings.Title("hello world")          // "Hello World" (deprecated)
cases.Title(language.English).String("hello world") // Use this instead

// Builder (efficient concatenation)
var b strings.Builder
for i := 0; i < 1000; i++ {
    b.WriteString("line\n")
}
result := b.String()

// Regular expressions
import "regexp"

re := regexp.MustCompile(`\d+`)
re.FindString("abc 123 def")          // "123"
re.FindAllString("abc 123 def 456", -1) // ["123", "456"]
re.MatchString("abc 123")             // true
re.ReplaceAllString("abc 123", "NUM") // "abc NUM"

// Submatches
re = regexp.MustCompile(`(\w+)@(\w+)\.(\w+)`)
matches := re.FindStringSubmatch("email: test@example.com")
// ["test@example.com", "test", "example", "com"]

// Named groups
re = regexp.MustCompile(`(?P<name>\w+)@(?P<domain>\w+)\.\w+`)
matches := re.FindStringSubmatch("test@example.com")
name := matches[re.SubexpIndex("name")]

// String conversion
import "strconv"

strconv.Itoa(42)                      // "42"
strconv.Atoi("42")                    // 42, nil
strconv.ParseFloat("3.14", 64)        // 3.14, nil
strconv.FormatFloat(3.14, 'f', 2, 64) // "3.14"
strconv.ParseBool("true")             // true, nil
strconv.Quote("hello")                // "\"hello\""
```

### 7. Sorting

```go
import "sort"

// Sort slice
ints := []int{3, 1, 4, 1, 5}
sort.Ints(ints)
sort.Sort(sort.Reverse(sort.IntSlice(ints)))

// Sort strings
strs := []string{"c", "a", "b"}
sort.Strings(strs)

// Sort by function
people := []Person{{"Bob", 30}, {"Alice", 25}}
sort.Slice(people, func(i, j int) bool {
    return people[i].Age < people[j].Age
})

// Sort with stability (Go 1.22+)
sort.SliceStable(people, func(i, j int) bool {
    return people[i].Age < people[j].Age
})

// Binary search
i := sort.SearchInts(ints, 3)
i := sort.Search(len(ints), func(i int) bool {
    return ints[i] >= 3
})

// Custom sort
type ByAge []Person

func (a ByAge) Len() int           { return len(a) }
func (a ByAge) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAge) Less(i, j int) bool { return a[i].Age < a[j].Age }

sort.Sort(ByAge(people))
```

### 8. Collections

```go
import "container/list"

// Doubly linked list
l := list.New()
l.PushBack(1)
l.PushFront(0)
l.PushBackList(otherList)

for e := l.Front(); e != nil; e = e.Next() {
    fmt.Println(e.Value)
}

// Heap
import "container/heap"

type IntHeap []int

func (h IntHeap) Len() int           { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *IntHeap) Push(x interface{}) { *h = append(*h, x.(int)) }
func (h *IntHeap) Pop() interface{} {
    old := *h
    n := len(old)
    x := old[n-1]
    *h = old[:n-1]
    return x
}

h := &IntHeap{2, 1, 5}
heap.Init(h)
heap.Push(h, 3)
min := heap.Pop(h)

// Ring
import "container/ring"

r := ring.New(5)
for i := 0; i < r.Len(); i++ {
    r.Value = i
    r = r.Next()
}
```

## Hands-on Exercises

Create the following programs in `phase-07-stdlib/`:

### Exercise 1: File Processor

Build a file processing tool that:

- Reads multiple file formats (JSON, CSV, XML)
- Transforms data
- Writes to different formats
- Handles large files with streaming

### Exercise 2: HTTP Client Library

Create a robust HTTP client that:

- Supports retries with backoff
- Has connection pooling
- Handles timeouts properly
- Logs requests/responses

### Exercise 3: Log Parser

Build a log parser that:

- Reads log files
- Parses timestamps
- Filters by date range
- Extracts specific patterns

### Exercise 4: Configuration Manager

Create a config system that:

- Reads from JSON/YAML/TOML files
- Supports environment variables
- Watches for file changes
- Validates configuration

## Resources

### Official

- [Package Documentation](https://pkg.go.dev/std)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)

### Deep Dives

- [Go Standard Library Cookbook](https://github.com/PacktPublishing/Go-Standard-Library-Cookbook)

## Validation Checklist

- [ ] Understand io interfaces
- [ ] Can work with files safely
- [ ] Can parse/generate JSON
- [ ] Can build HTTP clients
- [ ] Can build HTTP servers
- [ ] Can work with time
- [ ] Can use strings and regex
- [ ] All exercises completed

## Next Phase

Proceed to **Phase 08: Building Web Services** to build real-world web applications.
