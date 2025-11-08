MAKEFLAGS += --no-print-directory

DB_NAME := gormeasy_example


# Install git hooks
install-hooks:
	@chmod +x scripts/install-git-hooks.sh
	@./scripts/install-git-hooks.sh


lint:	
	@command -v gopls >/dev/null 2>&1 || { \
		echo "🔧 Installing gopls..."; \
		go install golang.org/x/tools/gopls@latest; \
	}
	@echo "🔍 Running gopls check..."
	@gopls check $$(find . -name '*.go' -type f -not -path "./ent/*" -not -path "./vendor/*")


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

check-db:
	@go run example/main.go test --test-db-name=migrate_test --owner-db-url=postgres://postgres:the_password@localhost:9433/postgres?sslmode=disable --test-db-url=postgres://postgres:the_password@localhost:9433/migrate_test?sslmode=disable

again:
	@make reset-db
	@make up
	@make gen
	@make status

