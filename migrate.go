package gormeasy

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// MigrationsHistory represents a record in the migrations table that tracks applied migrations.
// It stores the migration ID as the primary key.
type MigrationsHistory struct {
	ID string `gorm:"primaryKey"`
}

// TableName returns the name of the database table used to store migration history.
// It implements the gorm.Tabler interface to customize the table name.
func (MigrationsHistory) TableName() string {
	return "migrations"
}

// Migration is a type alias for gormigrate.Migration.
// It represents a single database migration with its ID, Up, and Down functions.
type Migration = gormigrate.Migration

func getMigrator(db *gorm.DB, migrations []*Migration) *gormigrate.Gormigrate {
	return gormigrate.New(db, &gormigrate.Options{
		TableName:                 "migrations",
		IDColumnName:              "id",
		IDColumnSize:              255,
		UseTransaction:            false, // Must disable transaction to prevent data loss during table recreation
		ValidateUnknownMigrations: true,
	}, migrations)
}

// RunMigrations executes migrations and compares the differences before and after execution.
func RunMigrations(db *gorm.DB, migrations []*Migration) error {
	if err := db.AutoMigrate(&MigrationsHistory{}); err != nil {
		return fmt.Errorf("failed to migrate migrations table: %w", err)
	}

	m := getMigrator(db, migrations)

	before := getAppliedIDs(db)

	fmt.Println("Running migrations...")

	if err := m.Migrate(); err != nil {
		return fmt.Errorf("migrate failed: %w", err)
	}

	after := getAppliedIDs(db)
	diff := findNewMigrations(before, after)

	if len(diff) == 0 {
		fmt.Println("✅ Migration complete (no change)")
		return nil
	}

	fmt.Println("✅ Migration complete.")
	fmt.Println("🆕 New migrations applied:")
	for _, id := range diff {
		fmt.Println("  -", id)
	}

	printMigrationStatus(db, migrations, false)
	return nil
}

// getAppliedIDs reads the set of migration IDs from the migrations table in the current database.
func getAppliedIDs(db *gorm.DB) map[string]bool {
	var applied []MigrationsHistory
	ids := make(map[string]bool)
	if err := db.Find(&applied).Error; err != nil {
		fmt.Println("Failed to read migration table:", err)
		return ids
	}
	for _, m := range applied {
		ids[m.ID] = true
	}
	return ids
}

// findNewMigrations returns the migration IDs that are new in after compared to before.
func findNewMigrations(before, after map[string]bool) []string {
	var diff []string
	for id := range after {
		if !before[id] {
			diff = append(diff, id)
		}
	}
	return diff
}

// printMigrationStatus prints the current migration status (Applied / Pending).
func printMigrationStatus(db *gorm.DB, migrations []*Migration, forcePrint bool) {
	if err := db.AutoMigrate(&MigrationsHistory{}); err != nil {
		fmt.Println("Failed to migrate migrations table:", err)
		return
	}
	applied := getAppliedIDs(db)

	appliedCount := 0
	pendingCount := 0
	for _, m := range migrations {
		if applied[m.ID] {
			appliedCount++
		} else {
			pendingCount++
		}
	}

	if appliedCount == len(migrations) && pendingCount == 0 && !forcePrint {
		fmt.Println("✅ All migrations are up to date.")
		return
	}

	fmt.Println("\n=== Migration Status ===")

	if appliedCount > 0 {
		fmt.Println("✅ Applied migrations:")
		for _, m := range migrations {
			if applied[m.ID] {
				fmt.Println("  -", m.ID)
			}
		}
	}

	if pendingCount > 0 {
		fmt.Println("\n❌ Pending migrations:")
		for _, m := range migrations {
			if !applied[m.ID] {
				fmt.Println("  -", m.ID)
			}
		}
	}

}

func rollbackAllMigrations(m *gormigrate.Gormigrate) error {
	for {
		if err := m.RollbackLast(); err != nil {
			if err == gormigrate.ErrNoRunMigration {
				break
			}
			return err
		}
	}
	return nil
}
