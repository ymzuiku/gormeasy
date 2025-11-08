package gormeasy

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

func DropTable(tx *gorm.DB, tableNames ...interface{}) error {
	for _, tableName := range tableNames {
		if reflect.TypeOf(tableName).Kind() != reflect.String {
			return fmt.Errorf("table name must be a string")
		}
		if hasTable := tx.Migrator().HasTable(tableName); !hasTable {
			return fmt.Errorf("table %s does not exist", tableName)
		}
	}
	return tx.Migrator().DropTable(tableNames...)
}

func CreateDatabase(db *gorm.DB, dbName string) error {
	var exists bool
	checkSQL := fmt.Sprintf("SELECT EXISTS(SELECT FROM pg_database WHERE datname = '%s')", dbName)
	if err := db.Raw(checkSQL).Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if !exists {
		createSQL := fmt.Sprintf("CREATE DATABASE \"%s\"", dbName)
		if err := db.Exec(createSQL).Error; err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		fmt.Printf("✅ Created database: %s\n", dbName)
	} else {
		fmt.Printf("⚠️  Database already exists: %s\n", dbName)
	}
	return nil
}

func DeleteDatabase(db *gorm.DB, dbName string) error {
	var exists bool
	checkSQL := fmt.Sprintf("SELECT EXISTS(SELECT FROM pg_database WHERE datname = '%s')", dbName)
	if err := db.Raw(checkSQL).Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if exists {
		// 终止所有连接再删除
		dropSQL := fmt.Sprintf(`DROP DATABASE "%s";`, dbName)

		if err := db.Exec(dropSQL).Error; err != nil {
			return fmt.Errorf("failed to delete database: %w", err)
		}
		fmt.Printf("🗑️  Deleted database: %s\n", dbName)
	} else {
		fmt.Printf("⚠️  Database does not exist: %s\n", dbName)
	}
	return nil
}
