package gormeasy

import (
	"fmt"
	"os"

	"gorm.io/gorm"
)

func getGorm(dbURL string, getDb func(string) (*gorm.DB, error)) (*gorm.DB, error) {
	url := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		url = dbURL
	}

	if url == "" {
		fmt.Println("database URL is required, please option one of the following:")
		fmt.Println("- easymigrate --db-url=postgres://postgres:the_password@localhost:5432/postgres?sslmode=disable")
		fmt.Println("- .env set DATABASE_URL=postgres://postgres:the_password@localhost:5432/postgres?sslmode=disable")
		os.Exit(1)
	}

	return getDb(url)
}
