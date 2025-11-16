package gormeasy

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// Params holds configuration parameters for database operations.
// It contains fields for action type, target, output path, and database connection URLs.
type Params struct {
	Action             string
	Target             string
	OutputPath         string
	DatabaseURL        string
	DevDatabaseURL     string
	TargetDatabaseName string
}

// Start initializes and runs the CLI application for managing database migrations.
// It loads environment variables from a .env file if present, sets up CLI commands for database operations,
// and handles command-line arguments. Supported commands include create-db, delete-db, up, down, gen, status, and regression.
// The migrations parameter should contain all migration definitions to be managed.
// The getGormFromURL function is used to create a GORM database connection from a connection URL string.
func Start(migrations []*Migration, getGormFromURL func(string) (*gorm.DB, error)) error {

	if err := godotenv.Load(); err != nil {
		// If .env file doesn't exist, just log warning and continue using environment variables
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	// If no arguments provided, silently return to allow the application to continue
	if len(os.Args) < 2 {
		return nil
	}

	command := os.Args[1]

	// Handle help
	if command == "help" || command == "--help" || command == "-h" {
		printHelp()
		os.Exit(0)
	}

	// Parse command-specific flags
	switch command {
	case "create-db":
		return handleCreateDB(getGormFromURL)
	case "delete-db":
		return handleDeleteDB(getGormFromURL)
	case "up":
		return handleUp(migrations, getGormFromURL)
	case "down":
		return handleDown(migrations, getGormFromURL)
	case "gen":
		return handleGen(getGormFromURL)
	case "status":
		return handleStatus(migrations, getGormFromURL)
	case "regression":
		return handleRegression(migrations, getGormFromURL)
	default:
		// Unknown command, silently return to allow the application to continue
		return nil
	}
}

func printHelp() {
	fmt.Println("easymigrate - Manage PostgreSQL databases and migrations")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create-db    Create a PostgreSQL database if it does not exist")
	fmt.Println("  delete-db    Delete a PostgreSQL database if it exists")
	fmt.Println("  up           Migrate the database up")
	fmt.Println("  down         Migrate the database down")
	fmt.Println("  gen          Generate GORM models from database")
	fmt.Println("  status       Show the current migration status")
	fmt.Println("  regression   Run regression test for all migrations and rollbacks")
	fmt.Println()
	fmt.Println("Use 'command -h' for command-specific help")
}

func handleCreateDB(getGormFromURL func(string) (*gorm.DB, error)) error {
	fs := flag.NewFlagSet("create-db", flag.ExitOnError)
	dbName := fs.String("db-name", "", "Name of the database to create")
	ownerDBURL := fs.String("owner-db-url", os.Getenv("OWNER_DATABASE_URL"), "Development database connection URL")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s create-db [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[2:])

	if *dbName == "" {
		return fmt.Errorf("db-name is required")
	}

	db, err := getGorm(*ownerDBURL, getGormFromURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := CreateDatabase(db, *dbName); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

func handleDeleteDB(getGormFromURL func(string) (*gorm.DB, error)) error {
	fs := flag.NewFlagSet("delete-db", flag.ExitOnError)
	dbName := fs.String("db-name", "", "Name of the database to delete")
	ownerDBURL := fs.String("owner-db-url", os.Getenv("OWNER_DATABASE_URL"), "Development database connection URL")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s delete-db [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[2:])

	if *dbName == "" {
		return fmt.Errorf("db-name is required")
	}
	if *ownerDBURL == "" {
		return fmt.Errorf("owner-db-url is required")
	}

	db, err := getGorm(*ownerDBURL, getGormFromURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := DeleteDatabase(db, *dbName); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

func handleUp(migrations []*Migration, getGormFromURL func(string) (*gorm.DB, error)) error {
	fs := flag.NewFlagSet("up", flag.ExitOnError)
	databaseURL := fs.String("db-url", os.Getenv("DATABASE_URL"), "Development database connection URL")
	noExit := fs.Bool("no-exit", false, "When success, do not exit")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s up [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[2:])

	db, err := getGorm(*databaseURL, getGormFromURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	err = RunMigrations(db, migrations)
	if err != nil {
		printMigrationStatus(db, migrations, false)
		return err
	}
	printMigrationStatus(db, migrations, false)
	if !*noExit {
		os.Exit(0)
	}
	return nil
}

func handleDown(migrations []*Migration, getGormFromURL func(string) (*gorm.DB, error)) error {
	fs := flag.NewFlagSet("down", flag.ExitOnError)
	databaseURL := fs.String("db-url", os.Getenv("DATABASE_URL"), "Development database connection URL")
	id := fs.String("id", "", "Rollback to specific migration ID")
	all := fs.Bool("all", false, "Rollback all migrations")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s down [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[2:])

	db, err := getGorm(*databaseURL, getGormFromURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	m := getMigrator(db, migrations)
	if *id != "" {
		if err := m.RollbackTo(*id); err != nil {
			printMigrationStatus(db, migrations, false)
			return fmt.Errorf("failed to rollback to migration: %w", err)
		}
		fmt.Printf("✅ Rollback to migration: %s complete.\n", *id)
	} else if *all {
		if err := rollbackAllMigrations(m); err != nil {
			printMigrationStatus(db, migrations, false)
			return fmt.Errorf("failed to rollback all migrations: %w", err)
		}
		fmt.Printf("✅ Rollback all migrations complete.\n")
	} else {
		if err := m.RollbackLast(); err != nil {
			printMigrationStatus(db, migrations, false)
			return fmt.Errorf("rollback failed: %w", err)
		}
		fmt.Println("✅ Rollback last complete.")
	}
	printMigrationStatus(db, migrations, false)
	os.Exit(0)
	return nil
}

func handleGen(getGormFromURL func(string) (*gorm.DB, error)) error {
	fs := flag.NewFlagSet("gen", flag.ExitOnError)
	databaseURL := fs.String("db-url", os.Getenv("DATABASE_URL"), "Development database connection URL")
	out := fs.String("out", "", "Output path for generated models")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s gen [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[2:])

	if *out == "" {
		return fmt.Errorf("out is required")
	}

	db, err := getGorm(*databaseURL, getGormFromURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	if err := generateGormCode(db, *out); err != nil {
		return fmt.Errorf("failed to generate GORM code: %w", err)
	}
	os.Exit(0)
	return nil
}

func handleStatus(migrations []*Migration, getGormFromURL func(string) (*gorm.DB, error)) error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	databaseURL := fs.String("db-url", os.Getenv("DATABASE_URL"), "Development database connection URL")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s status [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[2:])

	db, err := getGorm(*databaseURL, getGormFromURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	printMigrationStatus(db, migrations, false)
	os.Exit(0)
	return nil
}

func handleRegression(migrations []*Migration, getGormFromURL func(string) (*gorm.DB, error)) error {
	fs := flag.NewFlagSet("regression", flag.ExitOnError)
	ownerDatabaseURL := fs.String("owner-db-url", os.Getenv("OWNER_DATABASE_URL"), "Development database connection URL")
	devDatabaseURL := fs.String("regression-db-url", os.Getenv("REGRESSION_DATABASE_URL"), "Target database connection URL")
	regressionDatabaseName := fs.String("db-name", "", "Regression test database name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s regression [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}
	fs.Parse(os.Args[2:])

	if *ownerDatabaseURL == "" {
		return fmt.Errorf("owner-db-url is required")
	}

	if *devDatabaseURL == "" {
		return fmt.Errorf("regression-db-url is required")
	}

	if *regressionDatabaseName == "" {
		return fmt.Errorf("db-name is required")
	}

	ownerDB, err := getGorm(*ownerDatabaseURL, getGormFromURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	if err = DeleteDatabase(ownerDB, *regressionDatabaseName); err != nil {
		return err
	}
	if err = CreateDatabase(ownerDB, *regressionDatabaseName); err != nil {
		return err
	}

	devDB, err := getGorm(*devDatabaseURL, getGormFromURL)
	if err != nil {
		return err
	}
	m := getMigrator(devDB, migrations)

	if err = m.Migrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	printMigrationStatus(devDB, migrations, true)

	if err = rollbackAllMigrations(m); err != nil {
		return fmt.Errorf("failed to rollback all migrations: %w", err)
	}
	printMigrationStatus(devDB, migrations, true)

	if err = m.Migrate(); err != nil {
		return fmt.Errorf("failed to migrate again database: %w", err)
	}

	printMigrationStatus(devDB, migrations, true)

	fmt.Println("✅ Regression test complete, migration all up and all down, and migrate again, all pass.")

	os.Exit(0)
	return nil
}
