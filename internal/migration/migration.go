package migration

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"
)

func Migrate(db *sql.DB, migrationsDir string) error {
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Фильтруем и сортируем SQL файлы
	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	// Сортируем по имени (важно для порядка выполнения)
	sort.Strings(migrationFiles)

	// Выполняем каждый файл
	for _, filename := range migrationFiles {
		if err := executeMigrationFile(db, migrationsDir, filename); err != nil {
			return fmt.Errorf("migration failed in file %s: %w", filename, err)
		}
		log.Printf("Successfully executed: %s", filename)
	}

	return nil
}

func executeMigrationFile(db *sql.DB, dir, filename string) error {
	// Читаем содержимое файла
	filepath := filepath.Join(dir, filename)
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Разделяем SQL команды (если в файле несколько)
	queries := strings.Split(string(content), ";\n")

	for i, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}

		// Выполняем каждый запрос
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %d in %s: %w\nQuery: %s",
				i+1, filename, err, query)
		}
	}

	return nil
}
