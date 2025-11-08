package gormeasy

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/gen"
	"gorm.io/gorm"
)

// generateGormModel 从数据库结构反向生成 GORM model 文件
func generateGormCode(db *gorm.DB, basePath string) error {
	modelPath := filepath.Join(basePath)

	// 安全保护：防止误删项目根目录
	if basePath == "." || basePath == "/" {
		return fmt.Errorf("refusing to generate into critical directory: %s", basePath)
	}

	// 查询数据库所有表
	tables, err := db.Migrator().GetTables()
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}

	if err := clearDirectory(basePath); err != nil {
		return fmt.Errorf("failed to clear directory: %w", err)
	}

	fmt.Println("Generating GORM code for tables:", tables)

	// ======== 生成 model 层 ========
	gModel := gen.NewGenerator(gen.Config{
		OutPath:      modelPath,
		ModelPkgPath: "model",
		Mode:         gen.WithoutContext, // 纯结构体
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
