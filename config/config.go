package config

import (
	"os"
	"strings"
	"time"
)

var (
	LogDirs = getLogDirs()

	LokiURL = "http://localhost:3100/loki/api/v1/push"

	PollingInterval = 2 * time.Second
)

// Obtém os diretórios da variável de ambiente ou usa padrão
func getLogDirs() []string {
	envDirs := os.Getenv("LOG_DIRS") // Ex: "C:/Logs/PDV,C:/Logs/Emissor"
	if envDirs != "" {
		return strings.Split(envDirs, ",")
	}
	return []string{"C:/logs/pdv", "C:/logs/emissor"} // Padrão
}
