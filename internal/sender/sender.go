package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-softlogger/config"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Estrutura do stream de logs para Loki
type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// Estrutura do payload enviado para Loki
type LokiPayload struct {
	Streams []LokiStream `json:"streams"`
}

// SendToLoki envia logs para o Loki com tags personalizadas
func SendToLoki(filename, message, app string) {
	if message == "" {
		return
	}

	// Categoriza o tipo do Log
	level := "info"
	lowerMessage := strings.ToLower(message)

	if strings.Contains(lowerMessage, "erro") {
		level = "error"
	} else if strings.Contains(lowerMessage, "warning") || strings.Contains(lowerMessage, "aviso") || strings.Contains(lowerMessage, "atenção") {
		level = "warning"
	}

	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())

	// Formata a mensagem como JSON estruturado
	formattedMessage, err := json.Marshal(map[string]string{
		"level":     level,
		"message":   message,
		"file":      filename,
		"app":       app,
		"timestamp": time.Now().Format(time.RFC3339),
	})
	if err != nil {
		fmt.Println("Erro ao formatar mensagem:", err)
		return
	}

	// Construindo o payload
	payload := LokiPayload{
		Streams: []LokiStream{
			{
				Stream: map[string]string{
					"app":   app,
					"file":  filename,
					"level": level,
				},
				Values: [][]string{
					{timestamp, string(formattedMessage)},
				},
			},
		},
	}

	// Serializa para JSON
	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Erro ao serializar o payload:", err)
		return
	}

	// Envia o log para o Loki
	resp, err := http.Post(config.LokiURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Erro ao enviar log para Loki:", err)
		return
	}
	defer resp.Body.Close()

	// Lê a resposta do Loki para depuração
	respBody, _ := ioutil.ReadAll(resp.Body)

	// Exibe detalhes da resposta
	fmt.Println("Log enviado para Loki:", resp.Status)
	fmt.Printf("Enviado para Loki: %s\n", string(body))
	fmt.Printf("Resposta do Loki: %s\n", string(respBody))
}
