# Gorm Easy

A simple and easy-to-use database migration tool for GORM, built on top of [gormigrate](https://pkg.go.dev/github.com/go-gormigrate/gormigrate/v2@v2.1.5). Gorm Easy provides a CLI interface to manage database migrations with ease. It supports all databases that GORM supports, including PostgreSQL, MySQL, SQLite, SQL Server, and more.

## Features

- 🚀 Simple CLI interface for database migrations
- 📊 Migration status tracking
- 🔄 Rollback support (single, all, or to specific migration)
- 🗄️ Database creation and deletion
- 🤖 GORM model generation from database schema
- ✅ Migration testing utilities

## Installation

```bash
go get github.com/ymzuiku/gormeasy
```

## Quick Start

### 1. Setup Your Project

Create a main file that initializes Gorm Easy:

```go
package main

import (
    "log"
    "github.com/ymzuiku/gormeasy"
    "gorm.io/driver/postgres"  // or mysql, sqlite, sqlserver, etc.
    "gorm.io/gorm"
)

func getMigrations() []*gormeasy.Migration {
    return []*gormeasy.Migration{
        {
            ID: "20240101000000-create-users",
            Migrate: func(tx *gorm.DB) error {
                // Your migration logic here
                return tx.AutoMigrate(&User{})
            },
            Rollback: func(tx *gorm.DB) error {
                return gormeasy.DropTable(tx, "users")
            },
        },
    }
}

func main() {
    if err := gormeasy.Start(getMigrations(), func(url string) (*gorm.DB, error) {
        // Use the appropriate GORM driver for your database
        return gorm.Open(postgres.Open(url))  // For PostgreSQL
        // return gorm.Open(mysql.Open(url))  // For MySQL
        // return gorm.Open(sqlite.Open(url)) // For SQLite
    }); err != nil {
        log.Fatalf("failed to start gormeasy: %v", err)
    }
}
```

### 2. Configure Environment

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

Test all migrations by running them up, rolling back, and running them again. This command creates a test database, runs migrations, and verifies rollback functionality.

```bash
./your-app test \
  --owner-db-url postgres://user:password@localhost:5432/postgres \
  --test-db-url postgres://user:password@localhost:5432/testdb \
  --test-db-name testdb
```

**Flags:**

- `--owner-db-url` (required): Database connection URL with permissions to create databases (defaults to `DATABASE_URL` env var, or can use `DEV_DATABASE_URL` env var)
- `--test-db-url` (required): Target test database connection URL (can also use `TARGET_DATABASE_URL` env var)
- `--test-db-name` (required): Name of the test database to create

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

## Development

### Install Git Hooks

```bash
make install-hooks
```

### Update Dependencies

```bash
go mod tidy
go get -u ./...
go mod tidy
```

## License

See LICENSE file for details.
