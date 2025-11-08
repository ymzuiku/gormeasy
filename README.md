# Gorm Easy

It's used `https://pkg.go.dev/github.com/go-gormigrate/gormigrate/v2@v2.1.5`

## Example

create test database:

```sh
docker run --name pg --network=mynet -p 0.0.0.0:9433:5432 -e POSTGRES_PASSWORD=the_password -e PGDATA=/var/lib/postgresql/data/pgdata  -v ~/docker-data/postgres/data:/var/lib/postgresql/data -d --restart=always postgres:17
```

set .env:

```sh
# .env file:

DATABASE_URL=postgres://postgres:the_password@localhost:9433/gormeasy_example?sslmode=disable

```

update go required:

```sh
go mod tidy
go get -u ./...
go mod tidy
```

## Development

create git hooks:

```sh
make install-hooks
```
