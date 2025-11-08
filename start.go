package gormeasy

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

type Params struct {
	Action             string
	Target             string
	OutputPath         string
	DatabaseURL        string
	DevDatabaseURL     string
	TargetDatabaseName string
}

func Start(migrations []*Migration, getGormFromURL func(string) (*gorm.DB, error)) error {

	if err := godotenv.Load(); err != nil {
		// If .env file doesn't exist, just log warning and continue using environment variables
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	// Check if help is explicitly requested
	showHelp := false
	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" || arg == "help" {
			showHelp = true
			break
		}
	}

	app := &cli.App{
		Name:     "easymigrate",
		Usage:    "Manage PostgreSQL databases and migrations",
		HideHelp: !showHelp, // Only hide help if not explicitly requested
		Action: func(c *cli.Context) error {
			// When no command is provided, silently return to allow the application to continue
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "create-db",
				Usage: "Create a PostgreSQL database if it does not exist",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "db-name", Usage: "Name of the database to create", Required: true},
					&cli.StringFlag{Name: "owner-db-url", Usage: "Development database connection URL", Required: false, EnvVars: []string{"OWNER_DATABASE_URL"}},
				},
				Action: func(c *cli.Context) error {
					databaseURL := c.String("owner-db-url")
					dbName := c.String("db-name")
					db, err := getGorm(databaseURL, getGormFromURL)
					if err != nil {
						return fmt.Errorf("failed to open database: %w", err)
					}

					if err := CreateDatabase(db, dbName); err != nil {
						return err
					}

					os.Exit(0)

					return nil
				},
			},
			{
				Name:  "delete-db",
				Usage: "Delete a PostgreSQL database if it exists",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "db-name", Usage: "Name of the database to delete", Required: true},
					&cli.StringFlag{Name: "owner-db-url", Usage: "Development database connection URL", Required: true, EnvVars: []string{"OWNER_DATABASE_URL"}},
				},
				Action: func(c *cli.Context) error {

					databaseURL := c.String("owner-db-url")
					dbName := c.String("db-name")

					db, err := getGorm(databaseURL, getGormFromURL)
					if err != nil {
						return fmt.Errorf("failed to open database: %w", err)
					}

					if err := DeleteDatabase(db, dbName); err != nil {
						return err
					}

					os.Exit(0)

					return nil
				},
			},
			{
				Name:  "up",
				Usage: "Migrate the database up",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "db-url", Usage: "Development database connection URL", Required: false, EnvVars: []string{"DATABASE_URL"}},
					&cli.BoolFlag{Name: "no-exit", Usage: "When success, do not exit", Required: false},
				},
				Action: func(c *cli.Context) error {
					noExit := c.Bool("no-exit")
					databaseURL := c.String("db-url")

					db, err := getGorm(databaseURL, getGormFromURL)
					if err != nil {
						return fmt.Errorf("failed to open database: %w", err)
					}
					err = runMigrateWithDiff(db, migrations)
					if err != nil {
						printMigrationStatus(db, migrations, false)
						return err
					}
					printMigrationStatus(db, migrations, false)
					if !noExit {
						os.Exit(0)
					}
					return nil
				},
			},
			{
				Name:  "down",
				Usage: "Migrate the database down",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "db-url", Usage: "Development database connection URL", Required: false, EnvVars: []string{"DATABASE_URL"}},
					&cli.StringFlag{Name: "id", Usage: "Rollback to specific migration ID", Required: false},
					&cli.BoolFlag{Name: "all", Usage: "Rollback all migrations", Required: false},
				},
				Action: func(c *cli.Context) error {
					all := c.Bool("all")
					id := c.String("id")
					databaseURL := c.String("db-url")

					db, err := getGorm(databaseURL, getGormFromURL)
					if err != nil {
						return fmt.Errorf("failed to open database: %w", err)
					}
					m := getMigrator(db, migrations)
					if id != "" {
						if err := m.RollbackTo(id); err != nil {
							printMigrationStatus(db, migrations, false)
							return fmt.Errorf("failed to rollback to migration: %w", err)
						}
						fmt.Printf("✅ Rollback to migration: %s complete.\n", id)
					} else if all {
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
				},
			},
			{
				Name:  "gen",
				Usage: "Generate GORM models from database",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "db-url", Usage: "Development database connection URL", Required: false, EnvVars: []string{"DATABASE_URL"}},
					&cli.StringFlag{Name: "out", Usage: "Output path for generated models", Required: true},
				},
				Action: func(c *cli.Context) error {
					databaseURL := c.String("db-url")
					out := c.String("out")

					db, err := getGorm(databaseURL, getGormFromURL)
					if err != nil {
						return fmt.Errorf("failed to open database: %w", err)
					}
					if err := generateGormCode(db, out); err != nil {
						return fmt.Errorf("failed to generate GORM code: %w", err)
					}
					os.Exit(0)
					return nil
				},
			},
			{
				Name:  "status",
				Usage: "Show the current migration status",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "db-url", Usage: "Development database connection URL", Required: false, EnvVars: []string{"DATABASE_URL"}},
				},
				Action: func(c *cli.Context) error {
					databaseURL := c.String("db-url")

					db, err := getGorm(databaseURL, getGormFromURL)
					if err != nil {
						return fmt.Errorf("failed to open database: %w", err)
					}
					printMigrationStatus(db, migrations, false)
					os.Exit(0)
					return nil
				},
			},
			{
				Name:  "regression",
				Usage: "Run regression test for all migrations and rollbacks",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "owner-db-url", Usage: "Development database connection URL", Required: true, EnvVars: []string{"OWNER_DATABASE_URL"}},
					&cli.StringFlag{Name: "regression-db-url", Usage: "Target database connection URL", Required: true, EnvVars: []string{"REGRESSION_DATABASE_URL"}},
					&cli.StringFlag{Name: "db-name", Usage: "Regression test database name", Required: true},
				},
				Action: func(c *cli.Context) error {
					ownerDatabaseURL := c.String("owner-db-url")
					devDatabaseURL := c.String("regression-db-url")
					regressionDatabaseName := c.String("db-name")

					if ownerDatabaseURL == "" {
						return fmt.Errorf("owner-db-url is required")
					}

					if devDatabaseURL == "" {
						return fmt.Errorf("regression-db-url is required")
					}

					if regressionDatabaseName == "" {
						return fmt.Errorf("regression-db-name is required")
					}

					ownerDB, err := getGorm(ownerDatabaseURL, getGormFromURL)
					if err != nil {
						return fmt.Errorf("failed to open database: %w", err)
					}
					if err = DeleteDatabase(ownerDB, regressionDatabaseName); err != nil {
						return err
					}
					if err = CreateDatabase(ownerDB, regressionDatabaseName); err != nil {
						return err
					}

					devDB, err := getGorm(devDatabaseURL, getGormFromURL)
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
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		errStr := err.Error()
		// If help was explicitly requested, show help and exit
		if showHelp && strings.Contains(errStr, "help requested") {
			// Show help by running help command
			helpArgs := []string{os.Args[0], "help"}
			if helpErr := app.Run(helpArgs); helpErr != nil {
				// If help command also fails, just exit
				os.Exit(1)
			}
			os.Exit(0)
		}
		// Suppress help-related errors and unknown command errors when no command is provided
		if strings.Contains(errStr, "flag provided but not defined") ||
			strings.Contains(errStr, "No help topic") ||
			(strings.Contains(errStr, "command") && strings.Contains(errStr, "not found")) {
			// Silently return for unknown command errors (but not explicit help requests)
			return nil
		}
		fmt.Println("Error:", err)
		return err
	}

	// If help was requested and app.Run succeeded (help was shown), exit
	if showHelp {
		os.Exit(0)
	}

	return nil
}
