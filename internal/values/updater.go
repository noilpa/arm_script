package values

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Values struct {
	NodeSelector map[string]string `yaml:"nodeSelector,omitempty"`
	Tolerations  []Toleration      `yaml:"tolerations,omitempty"`
}

type Toleration struct {
	Key      string `yaml:"key"`
	Operator string `yaml:"operator"`
	Value    string `yaml:"value"`
	Effect   string `yaml:"effect"`
}

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

// checkAndAddFields проверяет и добавляет необходимые строки в файл values.yaml.
func checkAndAddFields(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Флаги, указывающие на наличие нужных строк в файле.
	nodeSelectorExists := false
	tolerationExists := false

	// Проверяем, есть ли нужные строки в файле.
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "nodeSelector:") {
			nodeSelectorExists = true
		}
		if strings.Contains(line, "tolerations") {
			tolerationExists = true
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	// Если обе строки уже присутствуют, файл не изменяется.
	if nodeSelectorExists && tolerationExists {
		return false, nil
	}

	// Открываем файл для добавления.
	file, err = os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Построение строк для добавления.
	var builder strings.Builder

	// Добавляем nodeSelector, если его нет.
	if !nodeSelectorExists {
		builder.WriteString("\nnodeSelector:\n  dedicated-to: multi-arch\n")
	}

	// Добавляем tolerations, если его нет.
	if !tolerationExists {
		builder.WriteString("tolerations:\n")
		builder.WriteString("  - key: dedicated-to\n")
		builder.WriteString("    operator: Equal\n")
		builder.WriteString("    value: multi-arch\n")
		builder.WriteString("    effect: NoSchedule\n")
	}

	// Записываем строки в конец файла.
	_, err = file.WriteString(builder.String())
	if err != nil {
		return false, err
	}

	return true, nil
}
