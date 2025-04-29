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

// Chat representa um chat retornado pela API
type Chat struct {
	ID string `json:"id"`
}

// Message representa uma mensagem retornada pela API
type Message struct {
	ChatID string `json:"chatId"`
	Time   string `json:"time"`
}

// MessageInfo armazena o ID do chat e o timestamp da última mensagem
type MessageInfo struct {
	ChatID string
	Time   string
}

// fetchChats faz a requisição HTTP para listar chats e retorna um slice de Chat
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
		return nil, fmt.Errorf("unmarshalling JSON chats: %w", err)
	}

	return chats, nil
}

// fetchChatMessages faz a requisição HTTP para obter mensagens de um chat específico
func fetchChatMessages(chatID, urlTemplate, apiKey string) ([]Message, error) {
	url := fmt.Sprintf(urlTemplate, chatID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request for messages: %w", err)
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-KEY", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making HTTP request for messages: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("unexpected status %d for messages: %s", res.StatusCode, string(body))
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body for messages: %w", err)
	}

	var msgs []Message
	if err := json.Unmarshal(body, &msgs); err != nil {
		return nil, fmt.Errorf("unmarshalling JSON messages: %w", err)
	}

	return msgs, nil
}

func main() {
	// Carrega variáveis de ambiente (.env)
	utils.LoadEnvVariables()

	apiKey := os.Getenv("X_API_KEY_OCTA")
	chatsURL := "https://artearena.api004.octadesk.services/chat?page=1&limit=100&sort[direction]=desc&sort[property]=createdAt"
	messagesURLTemplate := "https://artearena.api004.octadesk.services/chat/%s/messages"

	// 1) Obtém lista de chats
	chats, err := fetchChats(chatsURL, apiKey)
	if err != nil {
		log.Fatalf("Error fetching chats: %v", err)
	}

	// 2) Declara slice para armazenar tuplas (ChatID, Time)
	var messageList []MessageInfo

	// Opções de execução
	limit := 5 // número máximo de chats para testar
	count := 0
	printImmediate := true // imprime cada tupla assim que obtida

	// 3) Para cada chat, busca mensagens e anexa última
	for _, chat := range chats {
		if count >= limit {
			break
		}

		msgs, err := fetchChatMessages(chat.ID, messagesURLTemplate, apiKey)
		if err != nil {
			log.Printf("Error fetching messages for chat %s: %v", chat.ID, err)
			continue
		}

		if len(msgs) > 0 {
			last := msgs[len(msgs)-1]
			mi := MessageInfo{ChatID: chat.ID, Time: last.Time}
			messageList = append(messageList, mi)

			if printImmediate {
				// Imprime como tupla
				fmt.Printf("(%s, %s)\n", mi.ChatID, mi.Time)
			}
		}

		count++
	}

	// 4) Se não imprimir imediatamente, exibe lista completa ao final
	if !printImmediate {
		for _, mi := range messageList {
			fmt.Printf("(%s, %s)\n", mi.ChatID, mi.Time)

			// 5) Faz a requisição para o endpoint de exportação
			exportURL := "http://localhost:8080/v1/admin/octa/chat"
			exportReq, err := http.NewRequest("PUT", exportURL, nil)
			if err != nil {
				log.Fatalf("Error creating export request: %v", err)
			}
			exportReq.Header.Add("accept", "application/json")
			exportReq.Header.Add("X-Admin-Key", os.Getenv("ADMIN_KEY"))

			res, err := http.DefaultClient.Do(exportReq)
			if err != nil {
				log.Fatalf("Error making export request: %v", err)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				log.Fatalf("Error reading response body: %v", err)
			}

			fmt.Println(string(body))
		}
	}

}
