MAKEFLAGS += --no-print-directory

DB_NAME := gormeasy_example


RAW_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^[^0-9]*//')
VERSION := $(or $(RAW_TAG),0.0.0)
NEXT_VERSION := $(shell echo $(VERSION) | awk -F. '{printf "%d.%d.%d", $$1, $$2, $$3+1}')

.PHONY: tag
tag:
	@echo "Current version: $(VERSION)"
	@echo "Creating new version tag: v$(NEXT_VERSION)"
	git tag -a v$(NEXT_VERSION) -m "Release v$(NEXT_VERSION)"
	git push origin v$(NEXT_VERSION)
	@echo "✅ Tag v$(NEXT_VERSION) pushed to remote repository"

lint:	
	@command -v gopls >/dev/null 2>&1 || { \
		echo "🔧 Installing gopls..."; \
		go install golang.org/x/tools/gopls@latest; \
	}
	@echo "🔍 Running gopls check..."
	@gopls check $$(find . -name '*.go' -type f -not -path "./ent/*" -not -path "./vendor/*")
	@go test ./...


reset-db:
	@go run example/main.go delete-db --db-name=gormeasy_example --owner-db-url=postgres://postgres:the_password@localhost:9433/postgres?sslmode=disable
	@go run example/main.go create-db --db-name=gormeasy_example --owner-db-url=postgres://postgres:the_password@localhost:9433/postgres?sslmode=disable
	@echo "===> Done. Database gormeasy_example recreated successfully."

up:
	@go run example/main.go up
down:
	@go run example/main.go down
down-all:
	@go run example/main.go down --all
status:
	@go run example/main.go status
gen:
	@make up
	@go run example/main.go gen --out=generated/model

test:
	@go run example/main.go regression --db-name=migrate_test --owner-db-url=postgres://postgres:the_password@localhost:9433/postgres?sslmode=disable --regression-db-url=postgres://postgres:the_password@localhost:9433/migrate_test?sslmode=disable


again:
	@make reset-db
	@make up
	@make gen
	@make status

