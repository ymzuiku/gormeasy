# Gorm Easy

A simple and easy-to-use database migration tool for GORM, built on top of [gormigrate](https://pkg.go.dev/github.com/go-gormigrate/gormigrate/v2@v2.1.5). Gorm Easy provides a CLI interface to manage database migrations with ease. It supports all databases that GORM supports, including PostgreSQL, MySQL, SQLite, SQL Server, and more.

## Installation

Install Gorm Easy in your Go project:

```bash
go get github.com/ymzuiku/gormeasy
```

## Features

- 🚀 Simple CLI interface for database migrations
- 📊 Migration status tracking
- 🔄 Rollback support (single, all, or to specific migration)
- 🗄️ Database creation and deletion
- 🤖 GORM model generation from database schema
- ✅ Migration testing utilities

## Development Workflow

Gorm Easy follows a **database-first** development approach where migrations are the single source of truth for your database schema. Here's the complete workflow to get started:

### Project Setup

First, create a main file that initializes Gorm Easy:

```go
// main.go
package main

import (
    "log"
    "github.com/ymzuiku/gormeasy"
    "gorm.io/driver/postgres"  // or mysql, sqlite, sqlserver, etc.
    "gorm.io/gorm"
)

func main() {
    if err := gormeasy.Start(getMigrations(), func(url string) (*gorm.DB, error) {
        // Use the appropriate GORM driver for your database
        return gorm.Open(postgres.Open(url))  // For PostgreSQL
        // return gorm.Open(mysql.Open(url))  // For MySQL
        // return gorm.Open(sqlite.Open(url)) // For SQLite
    }); err != nil {
        log.Fatalf("failed to start gormeasy: %v", err)
    }

    // Your application code continues here after migrations
}
```

### Configure Environment

Create a `.env` file or set environment variables. The default environment variable is `DATABASE_URL`:

```bash
# For PostgreSQL
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable

# For MySQL
DATABASE_URL=user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local

# For SQLite
DATABASE_URL=sqlite.db
```

**Note:** You can also use `--db-url` flag to override the environment variable for specific commands.

### 1. Define Migrations (Single Source of Truth)

Create `internal/migration.go` as the **global unique data source** for your ORM. This file contains all your migration definitions:

```go
// internal/migration.go
package internal

import (
    "time"
    "github.com/ymzuiku/gormeasy"
    "gorm.io/gorm"
)

func GetMigrations() []*gormeasy.Migration {
    return []*gormeasy.Migration{
        {
            ID: "20240101000000-create-users",
            Migrate: func(tx *gorm.DB) error {
                // Define your schema changes here
                type User struct {
                    ID        uint      `gorm:"primaryKey"`
                    Name      string    `gorm:"type:varchar(100)"`
                    Email     string    `gorm:"type:varchar(255);uniqueIndex"`
                    CreatedAt time.Time
                    UpdatedAt time.Time
                }
                return tx.AutoMigrate(&User{})
            },
            Rollback: func(tx *gorm.DB) error {
                return gormeasy.DropTable(tx, "users")
            },
        },
        // Add more migrations...
    }
}
```

### 2. Run Database Migrations

Apply migrations to your database:

```bash
# Run all pending migrations
go run main.go up
```

This will:

- Execute all pending migrations from `internal/migration.go`
- Update the database schema
- Track applied migrations in the `migrations` table

### 3. Generate GORM Models from Database

After migrations are applied, generate GORM model structs from the actual database schema:

```bash
# Generate models from database to generated/model directory
go run main.go gen --out=generated/model
```

This command:

- Connects to your database
- Inspects the current schema
- Generates GORM model structs matching your database tables
- Saves them to `generated/model/` directory

**Important:** Always run `gen` after running `up` to keep your generated models in sync with the database.

### 4. Use Generated Models in Development

In your application code, import and use the generated models:

```go
// main.go or your service files
package main

import (
    "your-project/generated/model"
    "gorm.io/gorm"
)

func GetUserByEmail(db *gorm.DB, email string) (*model.User, error) {
    var user model.User
    err := db.Where("email = ?", email).First(&user).Error
    return &user, err
}
```

### Complete Workflow Example

```bash
# 1. Define your migration in internal/migration.go
# (Edit the file to add/modify migrations)

# 2. Apply migrations to database
go run main.go up

# 3. Generate GORM models from database
go run main.go gen --out=generated/model

# 4. Use generated models in your code
# (Import and use models from generated/model package)
```

### Workflow Benefits

- **Single Source of Truth**: `internal/migration.go` is the only place where you define schema changes
- **Type Safety**: Generated models ensure your Go code matches the database schema
- **Version Control**: Migrations are tracked and can be rolled back if needed
- **Team Collaboration**: Everyone follows the same migration → generate → use workflow

### Project Structure

```
your-project/
├── internal/
│   └── migration.go          # Single source of truth for schema
├── generated/
│   └── model/                # Auto-generated GORM models
│       ├── user.gen.go
│       └── order.gen.go
├── main.go                   # Your application entry point
└── .env                      # Database configuration
```

## Commands

### `create-db`

Create a database if it does not exist. **Note:** This command is primarily designed for PostgreSQL. For other databases, you may need to create databases manually.

```bash
./your-app create-db --db-name mydatabase --owner-db-url postgres://user:password@localhost:5432/postgres
```

**Flags:**

- `--db-name` (required): Name of the database to create
- `--owner-db-url` (optional): Database connection URL with permissions to create databases (defaults to `DATABASE_URL` env var, or can use `DEV_DATABASE_URL` env var)

### `delete-db`

Delete a database if it exists. **Note:** This command is primarily designed for PostgreSQL. For other databases, you may need to delete databases manually.

```bash
./your-app delete-db --db-name mydatabase --owner-db-url postgres://user:password@localhost:5432/postgres
```

**Flags:**

- `--db-name` (required): Name of the database to delete
- `--owner-db-url` (required): Database connection URL with permissions to delete databases (defaults to `DATABASE_URL` env var, or can use `DEV_DATABASE_URL` env var)

### `up`

Run all pending migrations.

```bash
./your-app up --db-url postgres://user:password@localhost:5432/dbname
```

**Flags:**

- `--db-url` (optional): Database connection URL (defaults to `DATABASE_URL` env var, or can use `DEV_DATABASE_URL` env var)
- `--no-exit` (optional): When successful, do not exit (useful for programmatic usage)

**Example:**

```bash
./your-app up
# Uses DATABASE_URL from environment by default
```

### `down`

Rollback migrations. By default, rolls back the last migration.

```bash
# Rollback last migration
./your-app down

# Rollback all migrations
./your-app down --all

# Rollback to specific migration ID
./your-app down --id 20240101000000-create-users
```

**Flags:**

- `--db-url` (optional): Database connection URL (defaults to `DATABASE_URL` env var, or can use `DEV_DATABASE_URL` env var)
- `--id` (optional): Rollback to specific migration ID
- `--all` (optional): Rollback all migrations

### `status`

Show the current migration status (applied and pending migrations).

```bash
./your-app status --db-url postgres://user:password@localhost:5432/dbname
```

**Flags:**

- `--db-url` (optional): Database connection URL (defaults to `DATABASE_URL` env var, or can use `DEV_DATABASE_URL` env var)

**Output:**

```
=== Migration Status ===
✅ Applied migrations:
  - 20240101000000-create-users
  - 20240102000000-create-orders

❌ Pending migrations:
  - 20240103000000-create-products
```

### `gen`

Generate GORM models from your database schema.

```bash
./your-app gen --out ./models --db-url postgres://user:password@localhost:5432/dbname
```

**Flags:**

- `--db-url` (optional): Database connection URL (defaults to `DATABASE_URL` env var, or can use `DEV_DATABASE_URL` env var)
- `--out` (required): Output path for generated models

### `test`

Test all migrations by running them in a specified test database. This command performs a complete migration cycle to verify that all migrations work correctly:

1. **Creates a test database** with the specified name (deletes it first if it exists)
2. **Runs all migrations** (first time)
3. **Rolls back all migrations**
4. **Runs all migrations again** (second time)

This ensures that:

- All migrations can be applied successfully
- All rollbacks work correctly
- Migrations can be re-applied after rollback
- The migration system is idempotent

```bash
./your-app test \
  --owner-db-url postgres://user:password@localhost:5432/postgres \
  --test-db-url postgres://user:password@localhost:5432/testdb \
  --test-db-name testdb
```

**Flags:**

- `--owner-db-url` (required): Database connection URL with permissions to create/delete databases (defaults to `DATABASE_URL` env var, or can use `DEV_DATABASE_URL` env var)
- `--test-db-url` (required): Target test database connection URL (can also use `TARGET_DATABASE_URL` env var)
- `--test-db-name` (required): Name of the test database to create and use for testing

**Example:**

```bash
# Test migrations in a dedicated test database
go run main.go test \
  --owner-db-url postgres://postgres:password@localhost:5432/postgres \
  --test-db-url postgres://postgres:password@localhost:5432/migration_test \
  --test-db-name migration_test
```

**What happens:**

1. The test database `migration_test` is deleted if it exists
2. A new `migration_test` database is created
3. All migrations from `internal/migration.go` are applied (first time)
4. Migration status is displayed
5. All migrations are rolled back
6. Migration status is displayed again
7. All migrations are applied again (second time)
8. Final migration status is displayed
9. Success message: "✅ Test complete, migration all up and all down, and migrate again, all pass."

**Use cases:**

- **CI/CD pipelines**: Automatically test migrations before deployment
- **Development**: Verify migrations work correctly before applying to production
- **Team collaboration**: Ensure all team members' migrations are compatible

## Example

See the `example/` directory for a complete working example.

### Running the Example

1. Start a PostgreSQL database:

```bash
docker run --name pg --network=mynet -p 0.0.0.0:9433:5432 \
  -e POSTGRES_PASSWORD=the_password \
  -e PGDATA=/var/lib/postgresql/data/pgdata \
  -v ~/docker-data/postgres/data:/var/lib/postgresql/data \
  -d --restart=always postgres:17
```

2. Set up `.env`:

```bash
DATABASE_URL=postgres://postgres:the_password@localhost:9433/gormeasy_example?sslmode=disable
```

3. Run migrations:

```bash
cd example
go run main.go up
```

### Running as a Service

You can combine migration commands with your application server. When `gormeasy.Start()` completes (e.g., when using `--no-exit` flag or when no command matches), your application code continues to execute. This allows you to:

1. Run migrations on startup
2. Start your HTTP server after migrations complete

**Example usage:**

```bash
# Run migrations and then start the server
go run example/main.go up --no-exit

# Or simply run without arguments to start the server directly
# (if no command matches, gormeasy.Start returns and your server code executes)
go run example/main.go
```

The example includes a simple HTTP server that starts after `gormeasy.Start()` completes. Visit `http://localhost:8080/ping` to test the server.

## Development

### Install Git Hooks

```bash
make install-hooks
```

### Update Dependencies

```bash
go get -u ./...
go mod tidy
```

## License

See LICENSE file for details.
