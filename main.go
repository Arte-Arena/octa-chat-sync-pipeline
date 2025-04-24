package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"script/utils"
)

type Chat struct {
	ID string `json:"id"`
}

func fetchChats(url, apiKey string) ([]Chat, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-KEY", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making HTTP request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", res.StatusCode, string(body))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var chats []Chat
	if err := json.Unmarshal(body, &chats); err != nil {
		return nil, fmt.Errorf("unmarshalling JSON: %w", err)
	}

	return chats, nil
}

func main() {
	utils.LoadEnvVariables()

	url := "https://artearena.api004.octadesk.services/chat?page=1&limit=100&sort[direction]=desc&sort[property]=createdAt"
	apiKey := os.Getenv("X_API_KEY_OCTA")

	chats, err := fetchChats(url, apiKey)
	if err != nil {
		log.Fatalf("Error fetching chats: %v", err)
	}

	fmt.Printf("IDs carregados (%d):\n", len(chats))
	for _, chat := range chats {
		fmt.Println("-", chat.ID)
	}
}
