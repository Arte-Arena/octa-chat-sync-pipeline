package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Ignora linhas vazias e comentários
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		// Divide a linha em chave e valor
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove aspas se existirem
		value = strings.Trim(value, `"'`)

		// Define a variável de ambiente
		os.Setenv(key, value)
	}

	return scanner.Err()
}

func main() {
	// Carrega variáveis do arquivo .env
	if err := loadEnv(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	url := "https://artearena.api004.octadesk.services/chat?page=1&limit=100&sort[direction]=desc&sort[property]=createdAt"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-KEY", os.Getenv("X_API_KEY_OCTA"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Println(string(body))
}
