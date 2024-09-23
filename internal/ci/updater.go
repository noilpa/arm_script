package ci

import (
	"bufio"
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Update(ctx context.Context, root string) ([]string, error) {
	log.Printf("Update ci in %s\n", root)

	var modifiedFiles []string

	// Поиск файла pipeline.yml внутри директории .github/workflows.
	searchPath := filepath.Join(root, ".github/workflows")

	// Поиск всех pipeline.yml в указанной директории.
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Проверка на соответствие файлу pipeline.yml.
		if strings.HasSuffix(path, "pipeline.yml") || strings.HasSuffix(path, "pipeline.yaml") {
			changed, err := checkAndAddBuildArch(path)
			if err != nil {
				return err
			}
			if changed {
				modifiedFiles = append(modifiedFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(modifiedFiles) > 0 {
		var builder strings.Builder
		builder.WriteString("Changed files:\n")
		for _, file := range modifiedFiles {
			builder.WriteString(file + "\n")
		}
		log.Println(builder.String())
	} else {
		log.Println("No changes.")
	}

	return modifiedFiles, nil
}

// checkAndAddBuildArch проверяет наличие строки `build_arch: amd64,arm64` в файле и добавляет ее, если отсутствует.
func checkAndAddBuildArch(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Флаги для проверки наличия нужных строк.
	buildArchExists := false

	// Сканируем файл и проверяем наличие нужной строки.
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		if strings.Contains(line, "build_arch: amd64,arm64") {
			buildArchExists = true
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	// Если строка уже присутствует, файл не изменяется.
	if buildArchExists {
		return false, nil
	}

	// Поиск места для добавления новой строки.
	inserted := false
	for i, line := range lines {
		if strings.Contains(line, "with:") {
			// Вставляем строку `build_arch: amd64,arm64` сразу после `with:`.
			lines = append(lines[:i+1], append([]string{"      build_arch: amd64,arm64"}, lines[i+1:]...)...)
			inserted = true
			break
		}
	}

	// Если `with:` не найден, добавляем `with:` и `build_arch` в конце блока `jobs:`.
	if !inserted {
		for i, line := range lines {
			if strings.Contains(line, "uses: InDriver/base-workflows/.github/workflows/go_pipeline.yaml@main") {
				lines = append(lines[:i+1], append([]string{"    with:", "      build_arch: amd64,arm64"}, lines[i+1:]...)...)
				break
			}
		}
	}

	// Открываем файл для записи обновленного содержимого.
	file, err = os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Записываем измененное содержимое обратно в файл.
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return false, err
		}
	}
	err = writer.Flush()
	if err != nil {
		return false, err
	}

	return true, nil
}
