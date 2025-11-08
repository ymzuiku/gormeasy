package gormeasy

import (
	"fmt"
	"reflect"
	"strings"

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
	dialectorName := db.Dialector.Name()

	switch dialectorName {
	case "postgres":
		return createPostgresDatabase(db, dbName)
	case "mysql":
		return createMySQLDatabase(db, dbName)
	case "sqlite":
		return fmt.Errorf("SQLite does not support CREATE DATABASE. SQLite uses file-based databases. Please create the database file manually or use a different database for this operation")
	default:
		return fmt.Errorf("database creation is not supported for %s. Currently supported: PostgreSQL, MySQL", dialectorName)
	}
}

func createPostgresDatabase(db *gorm.DB, dbName string) error {
	var exists bool
	// Escape single quotes in database name to prevent SQL injection
	escapedName := strings.ReplaceAll(dbName, "'", "''")
	checkSQL := fmt.Sprintf("SELECT EXISTS(SELECT FROM pg_database WHERE datname = '%s')", escapedName)
	if err := db.Raw(checkSQL).Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if !exists {
		escapedNameQuoted := fmt.Sprintf(`"%s"`, strings.ReplaceAll(dbName, `"`, `""`))
		createSQL := fmt.Sprintf("CREATE DATABASE %s", escapedNameQuoted)
		if err := db.Exec(createSQL).Error; err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		fmt.Printf("✅ Created database: %s\n", dbName)
	} else {
		fmt.Printf("⚠️  Database already exists: %s\n", dbName)
	}
	return nil
}

func createMySQLDatabase(db *gorm.DB, dbName string) error {
	var count int64
	// Escape backticks in database name
	escapedName := strings.ReplaceAll(dbName, "`", "``")
	checkSQL := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = '%s'", strings.ReplaceAll(escapedName, "'", "''"))
	if err := db.Raw(checkSQL).Scan(&count).Error; err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if count == 0 {
		escapedNameQuoted := fmt.Sprintf("`%s`", strings.ReplaceAll(dbName, "`", "``"))
		createSQL := fmt.Sprintf("CREATE DATABASE %s", escapedNameQuoted)
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
	dialectorName := db.Dialector.Name()

	switch dialectorName {
	case "postgres":
		return deletePostgresDatabase(db, dbName)
	case "mysql":
		return deleteMySQLDatabase(db, dbName)
	case "sqlite":
		return fmt.Errorf("SQLite does not support DROP DATABASE. SQLite uses file-based databases. Please delete the database file manually or use a different database for this operation")
	default:
		return fmt.Errorf("database deletion is not supported for %s. Currently supported: PostgreSQL, MySQL", dialectorName)
	}
}

func deletePostgresDatabase(db *gorm.DB, dbName string) error {
	var exists bool
	// Escape single quotes in database name to prevent SQL injection
	escapedName := strings.ReplaceAll(dbName, "'", "''")
	checkSQL := fmt.Sprintf("SELECT EXISTS(SELECT FROM pg_database WHERE datname = '%s')", escapedName)
	if err := db.Raw(checkSQL).Scan(&exists).Error; err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if exists {
		// Terminate all connections before dropping
		escapedNameQuoted := fmt.Sprintf(`"%s"`, strings.ReplaceAll(dbName, `"`, `""`))
		// First, terminate all connections to the database
		terminateSQL := fmt.Sprintf(`
			SELECT pg_terminate_backend(pg_stat_activity.pid)
			FROM pg_stat_activity
			WHERE pg_stat_activity.datname = '%s'
			AND pid <> pg_backend_pid();
		`, escapedName)
		_ = db.Exec(terminateSQL) // Ignore errors for termination

		dropSQL := fmt.Sprintf("DROP DATABASE %s", escapedNameQuoted)
		if err := db.Exec(dropSQL).Error; err != nil {
			return fmt.Errorf("failed to delete database: %w", err)
		}
		fmt.Printf("🗑️  Deleted database: %s\n", dbName)
	} else {
		fmt.Printf("⚠️  Database does not exist: %s\n", dbName)
	}
	return nil
}

func deleteMySQLDatabase(db *gorm.DB, dbName string) error {
	var count int64
	// Escape backticks in database name
	escapedName := strings.ReplaceAll(dbName, "`", "``")
	checkSQL := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = '%s'", strings.ReplaceAll(escapedName, "'", "''"))
	if err := db.Raw(checkSQL).Scan(&count).Error; err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if count > 0 {
		escapedNameQuoted := fmt.Sprintf("`%s`", strings.ReplaceAll(dbName, "`", "``"))
		dropSQL := fmt.Sprintf("DROP DATABASE %s", escapedNameQuoted)
		if err := db.Exec(dropSQL).Error; err != nil {
			return fmt.Errorf("failed to delete database: %w", err)
		}
		fmt.Printf("🗑️  Deleted database: %s\n", dbName)
	} else {
		fmt.Printf("⚠️  Database does not exist: %s\n", dbName)
	}
	return nil
}
