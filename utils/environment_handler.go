package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	ENV_PORT        = "PORT"
	ENV_MONGODB_URI = "MONGODB_URI"
	ENV             = "ENV"
	ADMIN_KEY       = "ADMIN_KEY"
	X_API_KEY_OCTA  = "X_API_KEY_OCTA"

	ENV_DEVELOPMENT = "development"
	ENV_RELEASE     = "production"
)

var allowedKeys = []string{ENV_PORT, ENV_MONGODB_URI, ENV, X_API_KEY_OCTA, ADMIN_KEY}

var allowedEnvValues = []string{ENV_DEVELOPMENT, ENV_RELEASE}

func LoadEnvVariables() {
	workDir, err := os.Getwd()
	if err != nil {
		panic("[ENV] Erro ao obter o diretório de trabalho: " + err.Error())
	}

	filePath := filepath.Join(workDir, ".env")

	file, err := os.Open(filePath)
	if err != nil {
		panic("[ENV] Erro ao abrir o arquivo .env: " + err.Error())
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		panic("[ENV] Erro ao obter informações do arquivo .env: " + err.Error())
	}

	if fileInfo.Size() == 0 {
		panic("[ENV] O arquivo .env está vazio")
	}

	foundKeys := make(map[string]bool)
	for _, key := range allowedKeys {
		foundKeys[key] = false
	}

	scanner := bufio.NewScanner(file)
	if err := scanner.Err(); err != nil {
		panic("[ENV] Erro ao criar scanner para o arquivo .env: " + err.Error())
	}

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			panic(fmt.Sprintf("[ENV] Formato inválido na linha %d: %s", lineNum, line))
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if len(value) > 1 && (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		if key == ENV {
			isValidEnv := slices.Contains(allowedEnvValues, value)

			if !isValidEnv {
				panic(fmt.Sprintf("[ENV] Valor inválido para ENV: %s. Valores permitidos: %s",
					value, strings.Join(allowedEnvValues, ", ")))
			}
		}

		isAllowed := slices.Contains(allowedKeys, key)

		if !isAllowed {
			panic(fmt.Sprintf("[ENV] Chave '%s' não é permitida. Chaves permitidas: %s",
				key, strings.Join(allowedKeys[:], ", ")))
		}

		if err := os.Setenv(key, value); err != nil {
			panic("[ENV] Erro ao definir variável de ambiente " + key + ": " + err.Error())
		}

		if _, exists := foundKeys[key]; exists {
			foundKeys[key] = true
		}
	}

	if err := scanner.Err(); err != nil {
		panic("[ENV] Erro ao ler o arquivo .env: " + err.Error())
	}

	var missingKeys []string
	for key, found := range foundKeys {
		if !found {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		panic(fmt.Sprintf("[ENV] Variáveis de ambiente obrigatórias ausentes: %s",
			strings.Join(missingKeys, ", ")))
	}
}
