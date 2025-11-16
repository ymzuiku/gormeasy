[English](README.md) | [中文](README-zh.md)

# Gorm Easy

一个简单易用的 GORM 数据库迁移工具，基于 [gormigrate](https://pkg.go.dev/github.com/go-gormigrate/gormigrate/v2) 构建。Gorm Easy 提供了 CLI 接口，可以轻松管理数据库迁移。它支持 GORM 支持的所有数据库，包括 PostgreSQL、MySQL、SQLite、SQL Server 等。

## 安装

在您的 Go 项目中安装 Gorm Easy：

```bash
go get github.com/ymzuiku/gormeasy
```

## 功能特性

- 🚀 简单的 CLI 接口用于数据库迁移
- 📊 迁移状态跟踪
- 🔄 回滚支持（单个、全部或回滚到指定迁移）
- 🗄️ 数据库创建和删除
- 🤖 从数据库架构生成 GORM 模型
- ✅ 迁移回归测试工具

## 开发工作流

Gorm Easy 遵循**数据库优先**的开发方法，其中迁移是数据库架构的唯一真实来源。以下是完整的入门工作流：

### 项目设置

首先，创建一个初始化 Gorm Easy 的主文件：

```go
// main.go
package main

import (
    "log"
    "github.com/ymzuiku/gormeasy"
    "gorm.io/driver/postgres"  // 或 mysql, sqlite, sqlserver 等
    "gorm.io/gorm"
    "internal/migration" // 我们将创建它
)

func main() {
    if err := gormeasy.Start(migration.GetMigrations(), func(url string) (*gorm.DB, error) {
        // 为您的数据库使用适当的 GORM 驱动
        return gorm.Open(postgres.Open(url))  // PostgreSQL
        // return gorm.Open(mysql.Open(url))  // MySQL
        // return gorm.Open(sqlite.Open(url)) // SQLite
    }); err != nil {
        log.Fatalf("failed to start gormeasy: %v", err)
    }

    // 迁移后您的应用程序代码继续执行
}
```

### 配置环境

创建 `.env` 文件或设置环境变量。默认环境变量是 `DATABASE_URL`：

```bash
# PostgreSQL
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable

# MySQL
DATABASE_URL=user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local

# SQLite
DATABASE_URL=sqlite.db
```

**注意：** 您也可以使用 `--db-url` 标志来覆盖特定命令的环境变量。

### 1. 定义迁移（单一真实来源）

创建 `internal/migration.go` 作为您的 ORM 的**全局唯一数据源**。此文件包含所有迁移定义：

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
                // 在此定义您的架构更改
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
        // 添加更多迁移...
    }
}
```

### 2. 运行数据库迁移

将迁移应用到您的数据库：

```bash
# 运行所有待处理的迁移
go run main.go up
```

这将：

- 从 `internal/migration.go` 执行所有待处理的迁移
- 更新数据库架构
- 在 `migrations` 表中跟踪已应用的迁移

### 3. 从数据库生成 GORM 模型

应用迁移后，从实际数据库架构生成 GORM 模型结构体：

```bash
# 从数据库生成模型到 generated/model 目录
go run main.go gen --out=generated/model
```

此命令：

- 连接到您的数据库
- 检查当前架构
- 生成与数据库表匹配的 GORM 模型结构体
- 将它们保存到 `generated/model/` 目录

**重要提示：** 运行 `up` 后始终运行 `gen`，以保持生成的模型与数据库同步。

### 4. 在开发中使用生成的模型

在您的应用程序代码中，导入并使用生成的模型：

```go
// main.go 或您的服务文件
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

### 完整工作流示例

```bash
# 1. 在 internal/migration.go 中定义您的迁移
# （编辑文件以添加/修改迁移）

# 2. 将迁移应用到数据库
go run main.go up

# 3. 从数据库生成 GORM 模型
go run main.go gen --out=generated/model

# 4. 在代码中使用生成的模型
# （从 generated/model 包导入并使用模型）
```

### 工作流优势

- **单一真实来源**：`internal/migration.go` 是您定义架构更改的唯一位置
- **类型安全**：生成的模型确保您的 Go 代码与数据库架构匹配
- **版本控制**：迁移被跟踪，如果需要可以回滚
- **团队协作**：每个人都遵循相同的迁移 → 生成 → 使用工作流

### 项目结构

```
your-project/
├── internal/
│   └── migration.go          # 架构的单一真实来源
├── generated/
│   └── model/                # 自动生成的 GORM 模型
│       ├── user.gen.go
│       └── order.gen.go
├── main.go                   # 您的应用程序入口点
└── .env                      # 数据库配置
```

## 命令

### `create-db`

如果数据库不存在则创建数据库。**注意：** 此命令主要设计用于 PostgreSQL。对于其他数据库，您可能需要手动创建数据库。

```bash
./your-app create-db --db-name mydatabase --owner-db-url postgres://user:password@localhost:5432/postgres
```

**标志：**

- `--db-name`（必需）：要创建的数据库名称
- `--owner-db-url`（可选）：具有创建数据库权限的数据库连接 URL（默认为 `OWNER_DATABASE_URL` 环境变量）

### `delete-db`

如果数据库存在则删除数据库。**注意：** 此命令主要设计用于 PostgreSQL。对于其他数据库，您可能需要手动删除数据库。

```bash
./your-app delete-db --db-name mydatabase --owner-db-url postgres://user:password@localhost:5432/postgres
```

**标志：**

- `--db-name`（必需）：要删除的数据库名称
- `--owner-db-url`（必需）：具有删除数据库权限的数据库连接 URL（默认为 `OWNER_DATABASE_URL` 环境变量）

### `up`

运行所有待处理的迁移。

```bash
./your-app up --db-url postgres://user:password@localhost:5432/dbname
```

**标志：**

- `--db-url`（可选）：数据库连接 URL（默认为 `DATABASE_URL` 环境变量）
- `--no-exit`（可选）：成功时不退出（对程序化使用很有用）

**示例：**

```bash
./your-app up
# 默认使用环境中的 DATABASE_URL
```

### `down`

回滚迁移。默认情况下，回滚最后一次迁移。

```bash
# 回滚最后一次迁移
./your-app down

# 回滚所有迁移
./your-app down --all

# 回滚到指定的迁移 ID
./your-app down --id 20240101000000-create-users
```

**标志：**

- `--db-url`（可选）：数据库连接 URL（默认为 `DATABASE_URL` 环境变量）
- `--id`（可选）：回滚到指定的迁移 ID
- `--all`（可选）：回滚所有迁移

### `status`

显示当前迁移状态（已应用和待处理的迁移）。

```bash
./your-app status --db-url postgres://user:password@localhost:5432/dbname
```

**标志：**

- `--db-url`（可选）：数据库连接 URL（默认为 `DATABASE_URL` 环境变量）

**输出：**

```
=== Migration Status ===
✅ Applied migrations:
  - 20240101000000-create-users
  - 20240102000000-create-orders

❌ Pending migrations:
  - 20240103000000-create-products
```

### `gen`

从数据库架构生成 GORM 模型。

```bash
./your-app gen --out ./models --db-url postgres://user:password@localhost:5432/dbname
```

**标志：**

- `--db-url`（可选）：数据库连接 URL（默认为 `DATABASE_URL` 环境变量）
- `--out`（必需）：生成模型的输出路径

### `regression`

通过在指定的测试数据库中运行所有迁移来执行回归测试。此命令执行完整的迁移周期以验证所有迁移是否正确工作：

1. **创建回归测试数据库**，使用指定的名称（如果存在则先删除）
2. **运行所有迁移**（第一次）
3. **回滚所有迁移**
4. **再次运行所有迁移**（第二次）

这确保了：

- 所有迁移都可以成功应用
- 所有回滚都正确工作
- 迁移可以在回滚后重新应用
- 迁移系统是幂等的

```bash
./your-app regression \
  --owner-db-url postgres://user:password@localhost:5432/postgres \
  --regression-db-url postgres://user:password@localhost:5432/regression_db \
  --db-name regression_db
```

**标志：**

- `--owner-db-url`（必需）：具有创建/删除数据库权限的数据库连接 URL（默认为 `OWNER_DATABASE_URL` 环境变量）
- `--regression-db-url`（必需）：目标回归测试数据库连接 URL（默认为 `REGRESSION_DATABASE_URL` 环境变量）
- `--db-name`（必需）：要创建并用于测试的回归测试数据库名称

**示例：**

```bash
# 在专用测试数据库中运行迁移的回归测试
go run main.go regression \
  --owner-db-url postgres://postgres:password@localhost:5432/postgres \
  --regression-db-url postgres://postgres:password@localhost:5432/migration_regression \
  --db-name migration_regression
```

**执行过程：**

1. 如果存在，则删除回归测试数据库 `migration_regression`
2. 创建新的 `migration_regression` 数据库
3. 从 `internal/migration.go` 应用所有迁移（第一次）
4. 显示迁移状态
5. 回滚所有迁移
6. 再次显示迁移状态
7. 再次应用所有迁移（第二次）
8. 显示最终迁移状态
9. 成功消息："✅ Regression test complete, migration all up and all down, and migrate again, all pass."

**使用场景：**

- **CI/CD 流水线**：在部署前自动测试迁移
- **开发**：在应用到生产环境之前验证迁移是否正确工作
- **团队协作**：确保所有团队成员的迁移兼容

## 示例

查看 `example/` 目录以获取完整的工作示例。

### 运行示例

1. 启动 PostgreSQL 数据库：

```bash
docker run --name pg --network=mynet -p 0.0.0.0:9433:5432 \
  -e POSTGRES_PASSWORD=the_password \
  -e PGDATA=/var/lib/postgresql/data/pgdata \
  -v ~/docker-data/postgres/data:/var/lib/postgresql/data \
  -d --restart=always postgres:17
```

2. 设置 `.env`：

```bash
DATABASE_URL=postgres://postgres:the_password@localhost:9433/gormeasy_example?sslmode=disable
```

3. 运行迁移：

```bash
cd example
go run main.go up
```

### 作为服务运行

您可以将迁移命令与应用程序服务器结合使用。当 `gormeasy.Start()` 完成时（例如，使用 `--no-exit` 标志或没有命令匹配时），您的应用程序代码将继续执行。这允许您：

1. 在启动时运行迁移
2. 在迁移完成后启动 HTTP 服务器

**使用示例：**

```bash
# 运行迁移然后启动服务器
go run example/main.go up --no-exit

# 或者简单地不带参数运行以直接启动服务器
# （如果没有命令匹配，gormeasy.Start 返回，您的服务器代码执行）
go run example/main.go
```

示例包含一个简单的 HTTP 服务器，在 `gormeasy.Start()` 完成后启动。访问 `http://localhost:8080/ping` 以测试服务器。

## 开发

### 安装 Git Hooks

```bash
make install-hooks
```

### 更新依赖

```bash
go get -u ./...
go mod tidy
```

## 依赖项

Gorm Easy 基于以下优秀的库构建：

### 核心库

- **[GORM](https://gorm.io/)** - 出色的 Go ORM 库，提供数据库抽象和查询构建
- **[GORM Gen](https://gorm.io/gen/)** - GORM 的代码生成工具，用于从数据库架构生成模型结构体
- **[gormigrate](https://github.com/go-gormigrate/gormigrate)** - GORM 的数据库迁移库，提供迁移管理和版本控制

### 支持库

- **[godotenv](https://github.com/joho/godotenv)** - 从 `.env` 文件加载环境变量
- **[GORM Drivers](https://gorm.io/docs/connecting_to_the_database.html)** - PostgreSQL、MySQL、SQLite、SQL Server 等的数据库驱动
- **Go 标准库 `flag`** - 命令行标志解析（内置，无外部依赖）

### 链接

- [gormigrate 文档](https://pkg.go.dev/github.com/go-gormigrate/gormigrate/v2)
- [GORM 文档](https://gorm.io/docs/)
- [GORM Gen 文档](https://gorm.io/gen/)

## 许可证

本项目采用 MIT 许可证。有关详细信息，请参阅 [LICENSE](LICENSE) 文件。
