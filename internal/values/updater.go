package values

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

func Update(ctx context.Context, root string) ([]string, error) {
	log.Printf("Update Values in %s\n", root)

	var modifiedFiles []string

	valuesDir := filepath.Join(root, "/deployments/aws")

	// Проверка существования директории /deployments/aws.
	if _, err := os.Stat(valuesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Dir %s not found.\n", valuesDir)
	}

	// Поиск всех *.values.yaml в указанной директории.
	err := filepath.Walk(valuesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Проверка пути на соответствие шаблону deployments/aws/*.values.yaml.
		if strings.HasSuffix(path, "values.yaml") {
			changed, err := checkAndAddFields(path)
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

// Структура для хранения данных из YAML
type YamlData map[string]interface{}

// checkAndAddFields проверяет и добавляет необходимые строки в файл values.yaml.
func checkAndAddFields(filePath string) (bool, error) {
	// Чтение содержимого файла
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %v", err)
	}

	// Парсинг YAML
	var yamlContent YamlData
	err = yaml.Unmarshal(fileData, &yamlContent)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal yaml: %v", err)
	}

	// Флаг изменений
	changed := false

	// Проверка и добавление nodeSelector
	if _, ok := yamlContent["nodeSelector"]; !ok {
		yamlContent["nodeSelector"] = map[string]string{
			"dedicated-to": "multi-arch",
		}
		changed = true
	}

	// Проверка и добавление tolerations
	if _, ok := yamlContent["tolerations"]; !ok {
		yamlContent["tolerations"] = []map[string]string{
			{
				"key":      "dedicated-to",
				"operator": "Equal",
				"value":    "multi-arch",
				"effect":   "NoSchedule",
			},
		}
		changed = true
	}

	// Проверка и модификация секции mysqlMigrations
	if mysqlMigrations, ok := yamlContent["mysqlMigrations"].(map[interface{}]interface{}); ok {
		if _, ok := mysqlMigrations["image"]; !ok {
			mysqlMigrations["image"] = "registry.idmp.tech/infra/liquibase-toolkit:0a626acd"
			changed = true
		}
	}

	// Если изменений не было, возвращаем false
	if !changed {
		return false, nil
	}

	// Сортировка корневых ключей для упорядоченного вывода
	keys := make([]string, 0, len(yamlContent))
	for k := range yamlContent {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Ручная сериализация каждого ключа с добавлением пустых строк между секциями
	var buffer bytes.Buffer
	for i, key := range keys {
		if i > 0 {
			buffer.WriteString("\n") // Добавляем пустую строку между секциями
		}

		// Сериализация каждого ключа отдельно
		sectionData := map[string]interface{}{key: yamlContent[key]}
		sectionYaml, err := yaml.Marshal(sectionData)
		if err != nil {
			return false, fmt.Errorf("failed to marshal yaml section for key %s: %v", key, err)
		}
		buffer.Write(sectionYaml)
	}

	// Запись обновленных данных обратно в файл
	err = os.WriteFile(filePath, buffer.Bytes(), 0644)
	if err != nil {
		return false, fmt.Errorf("failed to write updated yaml to file: %v", err)
	}

	// Возвращаем true, так как были изменения
	return true, nil
}
