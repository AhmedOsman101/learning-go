# Phase 09: Database and Persistence

**Duration**: 2-3 weeks  
**Prerequisites**: Phase 08 completed  
**Practice Directory**: `phase-09-databases/`

## Overview

Most applications need to persist data. This phase covers Go's database/sql package, working with relational databases, using ORMs, and handling migrations. You'll also learn about caching with Redis and NoSQL databases.

## Learning Objectives

- Master the database/sql package
- Work with connection pools
- Use prepared statements safely
- Implement transactions
- Work with popular ORMs (GORM, sqlc)
- Handle database migrations
- Integrate Redis for caching

## Topics to Cover

### 1. database/sql Package

```go
import (
    "database/sql"
    _ "github.com/lib/pq"  // PostgreSQL driver
)

// Connection
db, err := sql.Open("postgres", "postgres://user:pass@localhost/db?sslmode=disable")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Verify connection
if err := db.Ping(); err != nil {
    log.Fatal(err)
}

// Connection pool settings
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(10 * time.Minute)

// Query single row
var name string
err = db.QueryRow("SELECT name FROM users WHERE id = $1", 1).Scan(&name)

// Query multiple rows
rows, err := db.Query("SELECT id, name, email FROM users WHERE active = $1", true)
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

for rows.Next() {
    var id int
    var name, email string
    if err := rows.Scan(&id, &name, &email); err != nil {
        log.Fatal(err)
    }
    fmt.Println(id, name, email)
}

if err := rows.Err(); err != nil {
    log.Fatal(err)
}

// Insert
result, err := db.Exec(
    "INSERT INTO users (name, email) VALUES ($1, $2)",
    "John", "john@example.com",
)
if err != nil {
    log.Fatal(err)
}

id, _ := result.LastInsertId()  // Not supported by all drivers
rowsAffected, _ := result.RowsAffected()

// Update
result, err := db.Exec(
    "UPDATE users SET name = $1 WHERE id = $2",
    "Jane", 1,
)

// Delete
result, err := db.Exec("DELETE FROM users WHERE id = $1", 1)
```

### 2. Prepared Statements

```go
// Prepared statement (reusable)
stmt, err := db.Prepare("SELECT name FROM users WHERE id = $1")
if err != nil {
    log.Fatal(err)
}
defer stmt.Close()

// Use multiple times
var name string
err = stmt.QueryRow(1).Scan(&name)
err = stmt.QueryRow(2).Scan(&name)

// Named parameters (with sqlx)
import "github.com/jmoiron/sqlx"

db, _ := sqlx.Connect("postgres", dsn)

// Named query
user := User{Name: "John", Email: "john@example.com"}
result, err := db.NamedExec(
    "INSERT INTO users (name, email) VALUES (:name, :email)",
    user,
)

// Named query with struct
rows, err := db.NamedQuery(
    "SELECT * FROM users WHERE name = :name",
    map[string]interface{}{"name": "John"},
)

// Select into slice
var users []User
err = db.Select(&users, "SELECT * FROM users WHERE active = $1", true)

// Get single row
var user User
err = db.Get(&user, "SELECT * FROM users WHERE id = $1", 1)
```

### 3. Transactions

```go
// Basic transaction
tx, err := db.Begin()
if err != nil {
    log.Fatal(err)
}

// Rollback on error (safe to call after commit)
defer tx.Rollback()

// Execute within transaction
_, err = tx.Exec("INSERT INTO users (name) VALUES ($1)", "John")
if err != nil {
    return err  // defer will rollback
}

_, err = tx.Exec("INSERT INTO profiles (user_id) VALUES (currval('users_id_seq'))")
if err != nil {
    return err
}

// Commit
if err := tx.Commit(); err != nil {
    return err
}

// Transaction with context
tx, err := db.BeginTx(ctx, nil)

// Isolation levels
tx, err := db.BeginTx(ctx, &sql.TxOptions{
    Isolation: sql.LevelSerializable,
    ReadOnly:  false,
})

// Transaction pattern
func TransferFunds(db *sql.DB, fromID, toID int64, amount float64) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Check balance
    var balance float64
    err = tx.QueryRow(
        "SELECT balance FROM accounts WHERE id = $1 FOR UPDATE",
        fromID,
    ).Scan(&balance)
    if err != nil {
        return err
    }

    if balance < amount {
        return errors.New("insufficient funds")
    }

    // Debit
    _, err = tx.Exec(
        "UPDATE accounts SET balance = balance - $1 WHERE id = $2",
        amount, fromID,
    )
    if err != nil {
        return err
    }

    // Credit
    _, err = tx.Exec(
        "UPDATE accounts SET balance = balance + $1 WHERE id = $2",
        amount, toID,
    )
    if err != nil {
        return err
    }

    return tx.Commit()
}
```

### 4. NULL Handling

```go
// Nullable types
import "database/sql"

var (
    name sql.NullString
    age  sql.NullInt64
)

err := db.QueryRow("SELECT name, age FROM users WHERE id = $1", 1).Scan(&name, &age)

if name.Valid {
    fmt.Println(name.String)
}
if age.Valid {
    fmt.Println(age.Int64)
}

// Custom nullable type
type NullTime struct {
    Time  time.Time
    Valid bool
}

func (nt *NullTime) Scan(value interface{}) error {
    if value == nil {
        nt.Valid = false
        return nil
    }
    nt.Valid = true
    return nt.Time.Scan(value)
}

// Or use pointers
type User struct {
    ID        int64
    Name      string
    Email     *string  // NULL if nil
    CreatedAt time.Time
}

// sqlx null handling
import "github.com/jmoiron/sqlx/types"

type User struct {
    Name  string         `db:"name"`
    Email types.NullString `db:"email"`
}
```

### 5. GORM

```go
import "gorm.io/gorm"
import "gorm.io/driver/postgres"

// Model
type User struct {
    gorm.Model  // ID, CreatedAt, UpdatedAt, DeletedAt
    Name  string `gorm:"size:100;not null"`
    Email string `gorm:"size:255;uniqueIndex"`
    Posts []Post `gorm:"foreignKey:UserID"`
}

type Post struct {
    gorm.Model
    Title   string `gorm:"size:200"`
    Content string
    UserID  uint
    User    User
}

// Connection
dsn := "host=localhost user=postgres password=postgres dbname=test port=5432"
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

// Auto migrate
db.AutoMigrate(&User{}, &Post{})

// Create
user := User{Name: "John", Email: "john@example.com"}
result := db.Create(&user)
fmt.Println(user.ID)  // Filled after create

// Batch insert
users := []User{{Name: "A"}, {Name: "B"}}
db.Create(&users)

// Read
var user User
db.First(&user, 1)                    // By primary key
db.First(&user, "email = ?", "john@example.com")

// All
var users []User
db.Find(&users)

// Where
db.Where("name = ?", "John").First(&user)
db.Where("name <> ?", "John").Find(&users)
db.Where("name IN ?", []string{"John", "Jane"}).Find(&users)
db.Where("created_at > ?", yesterday).Find(&users)

// Order, Limit, Offset
db.Order("name desc").Limit(10).Offset(5).Find(&users)

// Select specific fields
db.Select("name", "email").Find(&users)

// Count
var count int64
db.Model(&User{}).Count(&count)

// Update
db.Model(&user).Update("name", "Jane")
db.Model(&user).Updates(User{Name: "Jane", Email: "jane@example.com"})
db.Model(&user).Updates(map[string]interface{}{"name": "Jane"})

// Delete
db.Delete(&user, 1)
db.Where("name = ?", "John").Delete(&User{})

// Soft delete (with gorm.Model)
db.Delete(&user)  // Sets DeletedAt
db.Unscoped().Find(&users)  // Include soft deleted

// Joins
db.Joins("JOIN posts ON posts.user_id = users.id").Find(&users)

// Preload
db.Preload("Posts").Find(&users)

// Transaction
err = db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    if err := tx.Create(&post).Error; err != nil {
        return err
    }
    return nil
})

// Raw SQL
var user User
db.Raw("SELECT * FROM users WHERE id = ?", 1).Scan(&user)

rows, _ := db.Raw("SELECT * FROM users").Rows()
defer rows.Close()
```

### 6. sqlc (Type-Safe SQL)

```go
// sqlc generates type-safe Go code from SQL queries

// 1. Write SQL queries
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY name;

-- name: CreateUser :one
INSERT INTO users (name, email) VALUES ($1, $2) RETURNING *;

-- name: UpdateUser :exec
UPDATE users SET name = $2 WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

// 2. Generate code
// sqlc generate

// 3. Use generated code
func main() {
    db, _ := sql.Open("postgres", dsn)
    queries := db.New(db)

    // Type-safe queries
    user, err := queries.GetUser(context.Background(), 1)
    users, err := queries.ListUsers(context.Background())

    createdUser, err := queries.CreateUser(context.Background(), db.CreateUserParams{
        Name:  "John",
        Email: "john@example.com",
    })

    err = queries.UpdateUser(context.Background(), db.UpdateUserParams{
        ID:   1,
        Name: "Jane",
    })
}
```

### 7. Migrations

```go
// golang-migrate
import "github.com/golang-migrate/migrate/v4"

// CLI usage
// migrate create -ext sql -dir migrations -seq create_users_table

// Migration file: 000001_create_users_table.up.sql
/*
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
*/

// Migration file: 000001_create_users_table.down.sql
/*
DROP TABLE users;
*/

// Run migrations
m, _ := migrate.New(
    "file://migrations",
    "postgres://user:pass@localhost:5432/db?sslmode=disable",
)
m.Up()
m.Down()
m.Steps(2)
m.Version()

// GORM migrations
type Migration struct {
    ID   string
    Up   func(*gorm.DB) error
    Down func(*gorm.DB) error
}

func Migrate(db *gorm.DB) error {
    migrations := []Migration{
        {
            ID: "202401010000",
            Up: func(db *gorm.DB) error {
                return db.AutoMigrate(&User{})
            },
            Down: func(db *gorm.DB) error {
                return db.Migrator().DropTable("users")
            },
        },
    }

    for _, m := range migrations {
        if err := m.Up(db); err != nil {
            return err
        }
    }
    return nil
}
```

### 8. Redis

```go
import "github.com/redis/go-redis/v9"

// Connection
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})

ctx := context.Background()

// Basic operations
rdb.Set(ctx, "key", "value", 0)
rdb.Set(ctx, "key", "value", 10*time.Minute)

val, err := rdb.Get(ctx, "key").Result()
if err == redis.Nil {
    // Key doesn't exist
}

rdb.Del(ctx, "key")
rdb.Exists(ctx, "key")

// Expiration
rdb.Expire(ctx, "key", time.Hour)
ttl := rdb.TTL(ctx, "key").Val()

// Increment
rdb.Incr(ctx, "counter")
rdb.IncrBy(ctx, "counter", 10)

// Lists
rdb.LPush(ctx, "list", "a", "b", "c")
rdb.RPush(ctx, "list", "d")
rdb.LPop(ctx, "list")
rdb.LRange(ctx, "list", 0, -1)

// Sets
rdb.SAdd(ctx, "set", "a", "b", "c")
rdb.SMembers(ctx, "set")
rdb.SIsMember(ctx, "set", "a")

// Hashes
rdb.HSet(ctx, "user:1", "name", "John", "email", "john@example.com")
rdb.HGet(ctx, "user:1", "name")
rdb.HGetAll(ctx, "user:1")

// Sorted sets
rdb.ZAdd(ctx, "leaderboard", &redis.Z{Score: 100, Member: "player1"})
rdb.ZRangeWithScores(ctx, "leaderboard", 0, -1)

// Pub/Sub
pubsub := rdb.Subscribe(ctx, "channel")
ch := pubsub.Channel()
for msg := range ch {
    fmt.Println(msg.Payload)
}

rdb.Publish(ctx, "channel", "message")

// Pipeline
pipe := rdb.Pipeline()
pipe.Set(ctx, "key1", "value1", 0)
pipe.Set(ctx, "key2", "value2", 0)
pipe.Get(ctx, "key1")
cmds, _ := pipe.Exec(ctx)

// Transaction
err := rdb.Watch(ctx, func(tx *redis.Tx) error {
    _, err := tx.TxPipeline().Exec(ctx)
    return err
}, "key")

// Cache pattern
func GetUserCached(ctx context.Context, rdb *redis.Client, db *sql.DB, id int) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", id)

    // Try cache first
    data, err := rdb.Get(ctx, cacheKey).Bytes()
    if err == nil {
        var user User
        json.Unmarshal(data, &user)
        return &user, nil
    }

    // Get from database
    user, err := GetUserFromDB(db, id)
    if err != nil {
        return nil, err
    }

    // Cache result
    data, _ = json.Marshal(user)
    rdb.Set(ctx, cacheKey, data, 5*time.Minute)

    return user, nil
}
```

## Hands-on Exercises

Create the following programs in `phase-09-databases/`:

### Exercise 1: User Repository

Implement a user repository with:

- CRUD operations
- Pagination
- Search/filter
- Soft delete

### Exercise 2: Transaction Handler

Build a transaction handler that:

- Supports nested transactions (savepoints)
- Has retry logic for deadlocks
- Logs all operations
- Times out properly

### Exercise 3: Caching Layer

Create a caching layer that:

- Caches query results
- Invalidates on updates
- Handles cache misses
- Supports multiple cache strategies

### Exercise 4: Migration System

Build a migration system that:

- Tracks applied migrations
- Supports up/down migrations
- Handles failures gracefully
- Can generate migration files

## Resources

### Official

- [database/sql documentation](https://pkg.go.dev/database/sql)
- [GORM](https://gorm.io/docs/)
- [sqlc](https://sqlc.dev/)

### Deep Dives

- [Go database/sql tutorial](http://go-database-sql.org/)
- [Golang Migrate](https://github.com/golang-migrate/migrate)

## Validation Checklist

- [ ] Can use database/sql package
- [ ] Understand connection pooling
- [ ] Can use prepared statements
- [ ] Can implement transactions
- [ ] Can use GORM or sqlc
- [ ] Can handle migrations
- [ ] Can use Redis for caching
- [ ] All exercises completed

## Next Phase

Proceed to **Phase 10: Testing and Benchmarking** to learn comprehensive testing strategies.
