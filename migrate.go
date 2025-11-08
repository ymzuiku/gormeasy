package gormeasy

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type migrationsModel struct {
	ID string `gorm:"primaryKey"`
}

type Migration = gormigrate.Migration

func getMigrator(db *gorm.DB, migrations []*Migration) *gormigrate.Gormigrate {
	return gormigrate.New(db, &gormigrate.Options{
		TableName:                 "migrations",
		IDColumnName:              "id",
		IDColumnSize:              255,
		UseTransaction:            false, // 必须关闭事务，防止表重建数据丢失
		ValidateUnknownMigrations: true,
	}, migrations)
}

// ============================================================
// 关键逻辑：执行前后对比迁移差异
// ============================================================
func runMigrateWithDiff(db *gorm.DB, migrations []*Migration) error {
	if err := db.AutoMigrate(&migrationsModel{}); err != nil {
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

// ============================================================
// 工具函数：查询迁移记录 + 差异对比
// ============================================================

// getAppliedIDs 读取当前数据库中 migrations 表的 ID 集合
func getAppliedIDs(db *gorm.DB) map[string]bool {
	var applied []migrationsModel
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

// findNewMigrations 返回 after 相比 before 新增的迁移 ID
func findNewMigrations(before, after map[string]bool) []string {
	var diff []string
	for id := range after {
		if !before[id] {
			diff = append(diff, id)
		}
	}
	return diff
}

// ============================================================
// 打印当前状态（Applied / Pending）
// ============================================================
func printMigrationStatus(db *gorm.DB, migrations []*Migration, forcePrint bool) {
	if err := db.AutoMigrate(&migrationsModel{}); err != nil {
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
