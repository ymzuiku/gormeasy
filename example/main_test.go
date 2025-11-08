package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// resetDatabase deletes and recreates the test database before each test
func resetDatabase(t *testing.T) {
	exampleDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	ownerDBURL := "postgres://postgres:the_password@localhost:9433/postgres?sslmode=disable"
	dbName := "gormeasy_example"

	// Delete database
	t.Logf("Deleting database %s...", dbName)
	deleteCmd := exec.Command("go", "run", "main.go", "delete-db", "--db-name="+dbName, "--owner-db-url="+ownerDBURL)
	deleteCmd.Dir = exampleDir
	deleteCmd.Stdout = os.Stdout
	deleteCmd.Stderr = os.Stderr
	_ = deleteCmd.Run() // Ignore errors, database might not exist

	// Create database
	t.Logf("Creating database %s...", dbName)
	createCmd := exec.Command("go", "run", "main.go", "create-db", "--db-name="+dbName, "--owner-db-url="+ownerDBURL)
	createCmd.Dir = exampleDir
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr
	if err := createCmd.Run(); err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
}

// TestMainUpCommandFirstRun tests the output of `go run example/main.go up` on first run
// This should show new migrations being applied
func TestMainUpCommandFirstRun(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set, skipping test")
	}

	// Reset database before test
	resetDatabase(t)

	// Get the directory of the test file (example directory)
	exampleDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Execute: go run main.go up (first time)
	cmd := exec.Command("go", "run", "main.go", "up")
	cmd.Dir = exampleDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	output := stdout.String() + stderr.String()

	// Check for expected output patterns for first run
	expectedPatterns := []string{
		"Running migrations",
		"Migration complete",
		"New migrations applied",
		"common-20251107100000-user",
		"common-20251107100000-order",
		"common-20251107100000-feedback",
		"All migrations are up to date",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(output, pattern) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nFull output:\n%s", pattern, output)
		}
	}

	// Verify the output contains migration success indicators
	if !strings.Contains(output, "✅") {
		t.Errorf("Expected output to contain success indicators (✅), but it didn't.\nFull output:\n%s", output)
	}

	// Verify it does NOT contain "no change" (that's for second run)
	if strings.Contains(output, "no change") {
		t.Errorf("First run should not contain 'no change'. Full output:\n%s", output)
	}

	// If command failed, show the error
	if err != nil {
		t.Logf("Command exited with error (this might be expected): %v", err)
		t.Logf("Full output:\n%s", output)
	}
}

// TestMainUpCommandSecondRun tests the output of `go run example/main.go up` on second run
// This should show "no change" since migrations are already applied
func TestMainUpCommandSecondRun(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set, skipping test")
	}

	// Reset database before test
	resetDatabase(t)

	exampleDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// First run: apply migrations
	t.Logf("First run: applying migrations...")
	firstCmd := exec.Command("go", "run", "main.go", "up")
	firstCmd.Dir = exampleDir
	_ = firstCmd.Run() // Ignore output for first run

	// Second run: should show "no change"
	t.Logf("Second run: should show no change...")
	secondCmd := exec.Command("go", "run", "main.go", "up")
	secondCmd.Dir = exampleDir

	var stdout, stderr bytes.Buffer
	secondCmd.Stdout = &stdout
	secondCmd.Stderr = &stderr

	err = secondCmd.Run()
	output := stdout.String() + stderr.String()

	// Check for expected output patterns for second run
	expectedPatterns := []string{
		"Running migrations",
		"Migration complete (no change)",
		"All migrations are up to date",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(output, pattern) {
			t.Errorf("Expected output to contain '%s', but it didn't.\nFull output:\n%s", pattern, output)
		}
	}

	// Verify it does NOT contain "New migrations applied" (that's for first run)
	if strings.Contains(output, "New migrations applied") {
		t.Errorf("Second run should not contain 'New migrations applied'. Full output:\n%s", output)
	}

	// Verify the output contains success indicators
	if !strings.Contains(output, "✅") {
		t.Errorf("Expected output to contain success indicators (✅), but it didn't.\nFull output:\n%s", output)
	}

	// If command failed, show the error
	if err != nil {
		t.Logf("Command exited with error (this might be expected): %v", err)
		t.Logf("Full output:\n%s", output)
	}
}

// TestMainUpCommandOutputFormat tests the specific format of the up command output
func TestMainUpCommandOutputFormat(t *testing.T) {
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set, skipping test")
	}

	// Reset database before test
	resetDatabase(t)

	exampleDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	cmd := exec.Command("go", "run", "main.go", "up")
	cmd.Dir = exampleDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	_ = cmd.Run()
	output := stdout.String() + stderr.String()

	// Check for specific output structure
	lines := strings.Split(output, "\n")

	hasRunningMigrations := false
	hasMigrationComplete := false
	hasAllUpToDate := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Running migrations") {
			hasRunningMigrations = true
		}
		if strings.Contains(line, "Migration complete") {
			hasMigrationComplete = true
		}
		if strings.Contains(line, "All migrations are up to date") {
			hasAllUpToDate = true
		}
	}

	// Verify expected output structure
	if !hasRunningMigrations {
		t.Errorf("Expected 'Running migrations...' in output")
	}
	if !hasMigrationComplete {
		t.Errorf("Expected 'Migration complete' in output")
	}
	if !hasAllUpToDate {
		t.Errorf("Expected 'All migrations are up to date' in output")
	}

	t.Logf("Output structure check passed. Full output:\n%s", output)
}

// TestMigrationsCount tests that getMigrations returns the expected number of migrations
func TestMigrationsCount(t *testing.T) {
	migrations := getMigrations()
	expectedCount := 3
	if len(migrations) != expectedCount {
		t.Errorf("Expected %d migrations, got %d", expectedCount, len(migrations))
	}

	// Verify migration IDs
	expectedIDs := []string{
		"common-20251107100000-user",
		"common-20251107100000-order",
		"common-20251107100000-feedback",
	}

	for i, expectedID := range expectedIDs {
		if i >= len(migrations) {
			t.Errorf("Migration %d not found", i)
			continue
		}
		if migrations[i].ID != expectedID {
			t.Errorf("Migration %d: expected ID '%s', got '%s'", i, expectedID, migrations[i].ID)
		}
	}
}

// TestMigrationStructure tests that migrations have the required structure
func TestMigrationStructure(t *testing.T) {
	migrations := getMigrations()

	for i, migration := range migrations {
		if migration.ID == "" {
			t.Errorf("Migration %d: ID is empty", i)
		}
		if migration.Migrate == nil {
			t.Errorf("Migration %d (%s): Migrate function is nil", i, migration.ID)
		}
		if migration.Rollback == nil {
			t.Errorf("Migration %d (%s): Rollback function is nil", i, migration.ID)
		}
	}
}
