package gormeasy

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/gen"
	"gorm.io/gorm"
)

// generateGormCode generates GORM model files by reverse engineering the database structure.
func generateGormCode(db *gorm.DB, basePath string) error {
	modelPath := filepath.Join(basePath)

	// Safety check: prevent accidental deletion of project root directory
	if basePath == "." || basePath == "/" {
		return fmt.Errorf("refusing to generate into critical directory: %s", basePath)
	}

	// Query all tables in the database
	tables, err := db.Migrator().GetTables()
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}

	if err := clearDirectory(basePath); err != nil {
		return fmt.Errorf("failed to clear directory: %w", err)
	}

	fmt.Println("Generating GORM code for tables:", tables)

	// Generate model layer
	gModel := gen.NewGenerator(gen.Config{
		OutPath:      modelPath,
		ModelPkgPath: "model",
		Mode:         gen.WithoutContext, // Pure structs only
	})
	gModel.UseDB(db)
	for _, table := range tables {
		gModel.GenerateModel(table)
	}
	gModel.Execute()
	fmt.Println("✅ Models generated in:", modelPath)

	fmt.Println("🎉 GORM code generation complete.")
	return nil
}

func clearDirectory(outputPath string) error {

	if outputPath == "" {
		return fmt.Errorf("missing output path, please set MODEL_DIR in .env file")
	}

	for _, p := range []string{outputPath} {
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("failed to clear dir %s: %w", p, err)
		}
		if err := os.MkdirAll(p, 0755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", p, err)
		}
	}
	return nil
}
