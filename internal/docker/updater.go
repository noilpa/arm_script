package docker

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Update(ctx context.Context, root string) ([]string, error) {
	log.Printf("Update Dockerfile in %s\n", root)

	buildDir := filepath.Join(root, "build")

	// Проверка существования директории /build.
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Dir %s not found.\n", buildDir)
	}

	// Поиск и модификация Dockerfile.
	modifiedFiles, err := findAndModifyDockerfiles(buildDir)
	if err != nil {
		return nil, fmt.Errorf("Error: %v\n", err)
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

// findAndModifyDockerfiles рекурсивно находит все Dockerfile в указанной директории и модифицирует их.
func findAndModifyDockerfiles(rootDir string) ([]string, error) {
	var modifiedFiles []string

	// Рекурсивный обход файлов.
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Проверка имени файла.
		if strings.ToLower(info.Name()) == "dockerfile" {
			changed, err := addPlatformOptionToDockerfile(path)
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
	return modifiedFiles, nil
}

// addPlatformOptionToDockerfile добавляет опцию --platform=${TARGETPLATFORM} к инструкциям FROM,
// если она отсутствует. Возвращает true, если файл был изменен.
func addPlatformOptionToDockerfile(filePath string) (bool, error) {
	// Открываем файл для чтения.
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	var lines []string
	modified := false
	scanner := bufio.NewScanner(file)

	// Чтение файла построчно и модификация инструкций FROM.
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "FROM") && !strings.Contains(line, "--platform=") {
			line = strings.Replace(line, "FROM", "FROM --platform=${TARGETPLATFORM}", 1)
			modified = true
		}
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	// Перезапись файла, если он был изменен.
	if modified {
		if err := os.WriteFile(filePath, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
			return false, err
		}
	}

	return modified, nil
}
