# Phase 08: Building Web Services

**Duration**: 2-3 weeks  
**Prerequisites**: Phase 07 completed  
**Practice Directory**: `phase-08-web-services/`

## Overview

This phase focuses on building production-ready web services in Go. You'll learn to structure applications, implement REST APIs, work with middleware, handle authentication, and integrate with popular frameworks.

## Learning Objectives

- Structure Go web applications properly
- Build RESTful APIs with best practices
- Implement middleware chains
- Handle authentication and authorization
- Work with popular frameworks (Gin, Echo, Chi)
- Implement WebSocket connections
- Build and consume gRPC services

## Topics to Cover

### 1. Project Structure

```
my-api/
├── cmd/
│   └── server/
│       └── main.go           # Entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration
│   ├── handler/
│   │   ├── user.go           # HTTP handlers
│   │   └── auth.go
│   ├── middleware/
│   │   ├── auth.go           # Middleware
│   │   └── logging.go
│   ├── service/
│   │   └── user.go           # Business logic
│   ├── repository/
│   │   └── user.go           # Data access
│   └── model/
│       └── user.go           # Domain models
├── pkg/
│   └── validator/
│       └── validator.go      # Shared utilities
├── api/
│   └── openapi.yaml          # API spec
├── migrations/
│   └── 001_create_users.sql  # DB migrations
├── go.mod
├── go.sum
├── Makefile
└── Dockerfile
```

### 2. RESTful API Design

```go
package main

import (
    "encoding/json"
    "net/http"
    "strconv"
)

// Handler with dependencies
type UserHandler struct {
    service UserService
}

func NewUserHandler(service UserService) *UserHandler {
    return &UserHandler{service: service}
}

// Request/Response DTOs
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

type UserResponse struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Handlers
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
    users, err := h.service.List(r.Context())
    if err != nil {
        respondError(w, http.StatusInternalServerError, err)
        return
    }
    respondJSON(w, http.StatusOK, users)
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
    if err != nil {
        respondError(w, http.StatusBadRequest, errors.New("invalid id"))
        return
    }

    user, err := h.service.Get(r.Context(), id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            respondError(w, http.StatusNotFound, err)
            return
        }
        respondError(w, http.StatusInternalServerError, err)
        return
    }

    respondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, err)
        return
    }

    if err := validate.Struct(req); err != nil {
        respondError(w, http.StatusBadRequest, err)
        return
    }

    user, err := h.service.Create(r.Context(), req)
    if err != nil {
        respondError(w, http.StatusInternalServerError, err)
        return
    }

    respondJSON(w, http.StatusCreated, user)
}

// Helpers
func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, code int, err error) {
    respondJSON(w, code, map[string]string{"error": err.Error()})
}

// Routing (Go 1.22+)
func main() {
    mux := http.NewServeMux()

    userHandler := NewUserHandler(userService)

    mux.HandleFunc("GET /api/users", userHandler.List)
    mux.HandleFunc("GET /api/users/{id}", userHandler.Get)
    mux.HandleFunc("POST /api/users", userHandler.Create)
    mux.HandleFunc("PUT /api/users/{id}", userHandler.Update)
    mux.HandleFunc("DELETE /api/users/{id}", userHandler.Delete)

    // Apply middleware
    handler := Chain(mux,
        LoggingMiddleware,
        RecoveryMiddleware,
        CORSMiddleware,
    )

    http.ListenAndServe(":8080", handler)
}
```

### 3. Middleware

```go
// Middleware type
type Middleware func(http.Handler) http.Handler

// Chain middleware
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}

// Logging middleware
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Wrap response writer to capture status
        wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

        next.ServeHTTP(wrapped, r)

        log.Printf("%s %s %d %v",
            r.Method,
            r.URL.Path,
            wrapped.status,
            time.Since(start),
        )
    })
}

// Response writer wrapper
type responseWriter struct {
    http.ResponseWriter
    status int
}

func (w *responseWriter) WriteHeader(status int) {
    w.status = status
    w.ResponseWriter.WriteHeader(status)
}

// Recovery middleware
func RecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %v", err)
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// CORS middleware
func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}

// Request ID middleware
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        ctx := context.WithValue(r.Context(), "requestID", requestID)
        w.Header().Set("X-Request-ID", requestID)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Rate limiting middleware
func RateLimitMiddleware(limiter *rate.Limiter) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Timeout middleware
func TimeoutMiddleware(timeout time.Duration) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, cancel := context.WithTimeout(r.Context(), timeout)
            defer cancel()

            done := make(chan struct{})
            go func() {
                next.ServeHTTP(w, r.WithContext(ctx))
                close(done)
            }()

            select {
            case <-done:
            case <-ctx.Done():
                http.Error(w, "request timeout", http.StatusRequestTimeout)
            }
        })
    }
}
```

### 4. Authentication

```go
import (
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

// Password hashing
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// JWT
type Claims struct {
    UserID int64  `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

func GenerateToken(userID int64, role, secret string) (string, error) {
    claims := Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

func ParseToken(tokenString, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}

// Auth middleware
func AuthMiddleware(secret string) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }

            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            claims, err := ParseToken(tokenString, secret)
            if err != nil {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }

            ctx := context.WithValue(r.Context(), "user", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Role-based access
func RequireRole(role string) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims, ok := r.Context().Value("user").(*Claims)
            if !ok || claims.Role != role {
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### 5. Using Frameworks

```go
// Gin framework
import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()  // Logger + Recovery middleware

    // Routes
    r.GET("/users", listUsers)
    r.GET("/users/:id", getUser)
    r.POST("/users", createUser)

    // Group with middleware
    auth := r.Group("/api")
    auth.Use(AuthMiddleware("secret"))
    {
        auth.GET("/profile", getProfile)
        auth.PUT("/profile", updateProfile)
    }

    r.Run(":8080")
}

func getUser(c *gin.Context) {
    id := c.Param("id")

    // Bind JSON
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Bind query
    var query QueryParams
    c.ShouldBindQuery(&query)

    // Response
    c.JSON(http.StatusOK, gin.H{"user": user})
}

// Echo framework
import "github.com/labstack/echo/v4"

func main() {
    e := echo.New()

    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())

    // Routes
    e.GET("/users", listUsers)
    e.GET("/users/:id", getUser)
    e.POST("/users", createUser)

    e.Logger.Fatal(e.Start(":8080"))
}

func getUser(c echo.Context) error {
    id := c.Param("id")
    return c.JSON(http.StatusOK, user)
}

// Chi router (lightweight, stdlib-compatible)
import "github.com/go-chi/chi/v5"

func main() {
    r := chi.NewRouter()

    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    r.Route("/users", func(r chi.Router) {
        r.Get("/", listUsers)
        r.Post("/", createUser)

        r.Route("/{id}", func(r chi.Router) {
            r.Get("/", getUser)
            r.Put("/", updateUser)
            r.Delete("/", deleteUser)
        })
    })

    http.ListenAndServe(":8080", r)
}
```

### 6. WebSocket

```go
import "github.com/gorilla/websocket"

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true  // Configure appropriately
    },
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    defer conn.Close()

    for {
        messageType, message, err := conn.ReadMessage()
        if err != nil {
            log.Println(err)
            break
        }

        log.Printf("received: %s", message)

        // Echo back
        err = conn.WriteMessage(messageType, message)
        if err != nil {
            log.Println(err)
            break
        }
    }
}

// Chat room example
type ChatRoom struct {
    clients   map[*websocket.Conn]bool
    broadcast chan []byte
    register  chan *websocket.Conn
    mutex     sync.Mutex
}

func (r *ChatRoom) Run() {
    for {
        select {
        case conn := <-r.register:
            r.mutex.Lock()
            r.clients[conn] = true
            r.mutex.Unlock()

        case message := <-r.broadcast:
            r.mutex.Lock()
            for client := range r.clients {
                err := client.WriteMessage(websocket.TextMessage, message)
                if err != nil {
                    client.Close()
                    delete(r.clients, client)
                }
            }
            r.mutex.Unlock()
        }
    }
}
```

### 7. gRPC

```go
// Proto file (user.proto)
/*
syntax = "proto3";

package user;

service UserService {
    rpc Get(GetRequest) returns (User);
    rpc List(ListRequest) returns (stream User);
    rpc Create(CreateRequest) returns (User);
}

message User {
    int64 id = 1;
    string name = 2;
    string email = 3;
}

message GetRequest {
    int64 id = 1;
}

message ListRequest {}

message CreateRequest {
    string name = 1;
    string email = 2;
}
*/

// Server
import (
    "google.golang.org/grpc"
    pb "path/to/generated/proto"
)

type server struct {
    pb.UnimplementedUserServiceServer
    // dependencies
}

func (s *server) Get(ctx context.Context, req *pb.GetRequest) (*pb.User, error) {
    user, err := s.service.Get(ctx, req.Id)
    if err != nil {
        return nil, err
    }
    return &pb.User{
        Id:    user.ID,
        Name:  user.Name,
        Email: user.Email,
    }, nil
}

func main() {
    lis, _ := net.Listen("tcp", ":50051")
    s := grpc.NewServer()
    pb.RegisterUserServiceServer(s, &server{})
    s.Serve(lis)
}

// Client
func main() {
    conn, _ := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
    defer conn.Close()

    client := pb.NewUserServiceClient(conn)

    user, _ := client.Get(context.Background(), &pb.GetRequest{Id: 1})
    fmt.Println(user)
}
```

## Hands-on Exercises

Create the following programs in `phase-08-web-services/`:

### Exercise 1: REST API

Build a complete REST API for a blog:

- CRUD for posts and comments
- Authentication with JWT
- Input validation
- Pagination and filtering

### Exercise 2: Authentication Service

Create an auth service with:

- User registration/login
- JWT token generation
- Refresh tokens
- Password reset

### Exercise 3: Real-time Chat

Build a WebSocket chat application:

- Multiple chat rooms
- User presence
- Message history
- Typing indicators

### Exercise 4: gRPC Microservice

Create a gRPC service that:

- Has multiple RPC methods
- Uses streaming
- Has proper error handling
- Includes a REST gateway

## Resources

### Official

- [net/http documentation](https://pkg.go.dev/net/http)
- [gRPC Go Quick Start](https://grpc.io/docs/languages/go/quickstart/)

### Frameworks

- [Gin](https://gin-gonic.com/)
- [Echo](https://echo.labstack.com/)
- [Chi](https://github.com/go-chi/chi)

## Validation Checklist

- [ ] Can structure a Go web project
- [ ] Can build RESTful APIs
- [ ] Can write middleware
- [ ] Can implement JWT auth
- [ ] Can use at least one framework
- [ ] Can work with WebSockets
- [ ] Understand gRPC basics
- [ ] All exercises completed

## Next Phase

Proceed to **Phase 09: Database and Persistence** to learn database integration.
