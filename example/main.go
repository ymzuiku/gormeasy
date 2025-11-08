package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ymzuiku/gormeasy"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getMigrations() []*gormeasy.Migration {
	return []*gormeasy.Migration{
		{
			ID: "common-20251107100000-user",
			Migrate: func(tx *gorm.DB) error {
				type user struct {
					ID        string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
					CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;index"`
					UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime;index"`
					Name      string    `json:"name" gorm:"type:varchar(64);index"`
					Email     string    `json:"email" gorm:"type:varchar(255);unique"`
					Role      string    `json:"role" gorm:"type:varchar(64);default:'customer'"`
				}

				return tx.AutoMigrate(&user{})

			},
			Rollback: func(tx *gorm.DB) error {
				return gormeasy.DropTable(tx, "users")
			},
		},
		{
			ID: "common-20251107100000-order",
			Migrate: func(tx *gorm.DB) error {

				type order struct {
					ID       string `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
					UserID   string `json:"user_id" gorm:"type:varchar(255);index"`
					Amount   int    `json:"amount" gorm:"type:integer;default:0"`
					Currency string `json:"currency" gorm:"type:varchar(10);default:'usd'"`
				}

				return tx.AutoMigrate(&order{})

			},
			Rollback: func(tx *gorm.DB) error {
				return gormeasy.DropTable(tx, "orders")
			},
		},
		{
			ID: "common-20251107100000-feedback",
			Migrate: func(tx *gorm.DB) error {

				type feedback struct {
					ID     string `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
					UserID string `json:"user_id" gorm:"type:varchar(255);index"`
					Title  string `json:"title" gorm:"type:varchar(255);index"`
					Rating int    `json:"rating" gorm:"type:integer;default:0"`
				}

				type feedbackContent struct {
					ID         string `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
					FeedbackID string `json:"feedback_id" gorm:"type:varchar(255);index"`
					Content    string `json:"content" gorm:"type:varchar(4096)"`
				}

				return tx.AutoMigrate(&feedback{}, &feedbackContent{})

			},
			Rollback: func(tx *gorm.DB) error {
				return gormeasy.DropTable(tx, "feedbacks", "feedback_contents")
			},
		},
	}

}

func main() {
	if err := gormeasy.Start(getMigrations(), func(url string) (*gorm.DB, error) {
		return gorm.Open(postgres.Open(url))
	}); err != nil {
		log.Fatalf("failed to start gormeasy: %v", err)
	}

	// Start HTTP server after gormeasy commands
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	})

	log.Println("Server starting on :8080")
	log.Println("Visit http://localhost:8080/ping to test")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
