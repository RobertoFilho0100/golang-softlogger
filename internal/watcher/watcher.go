package watcher

import (
	"bufio"
	"fmt"
	"go-softlogger/internal/sender"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

const debounceInterval = 2 * time.Second

var lastSent map[string]time.Time

func WatchLogs(logDirs []string) {
	// Inicializa o mapa de últimos envios
	lastSent = make(map[string]time.Time)

	// Cria o watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Erro ao iniciar watcher:", err)
		return
	}
	defer watcher.Close()

	// Adiciona os diretórios e arquivos de log para monitoramento
	for _, dir := range logDirs {
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".txt") {
				normalizedFilename := normalizeFilename(path)
				fmt.Println("Monitorando: ", normalizedFilename)
				watcher.Add(normalizedFilename)
			}
			return nil
		})
	}

	// Processa eventos de modificação de arquivos
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				// Para cada arquivo modificado, processa a última linha
				normalizedFilename := normalizeFilename(event.Name)
				fmt.Println("Arquivo modificado:", normalizedFilename)

				//A obtenção da última linha do log
				lastLine := getLastLines(event.Name, 1)
				if lastLine != "" {
					// Obtemos o nome da aplicação com base no diretório
					appName := getApplicationName(event.Name)

					if canSendLog(event.Name) {
						// Envia o log para Loki com as tags apropriadas
						sender.SendToLoki(event.Name, lastLine, appName)
					}
				}

				//lastLog := getLastLogEntry(normalizedFilename)
				//if lastLog != "" {
				//	appName := getApplicationName(normalizedFilename)
				//
				//	if canSendLog(normalizedFilename) {
				//		sender.SendToLoki(normalizedFilename, lastLog, appName)
				//	}
				//}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("Erro no watcher:", err)
		}
	}
}

func getLastLogEntry(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Erro ao abrir arquivo:", err)
		return ""
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue // Ignora linhas vazias
		}

		if isNewLogEntry(line) {
			lines = []string{line} // Começa nova entrada de log
		} else {
			lines = append(lines, line) // Continua a entrada atual
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Erro ao ler arquivo:", err)
		return ""
	}

	return strings.Join(lines, "\n") // Retorna a última entrada de log completa
}

func isNewLogEntry(line string) bool {
	// Supondo que cada entrada de log começa com uma data no formato [YYYY-MM-DD HH:MM:SS]
	return strings.HasPrefix(line, "[")
}

// Função para verificar se podemos enviar o log (evita o envio excessivo)
func canSendLog(filename string) bool {
	now := time.Now()
	if lastTime, exists := lastSent[filename]; exists {
		if now.Sub(lastTime) < debounceInterval {
			return false // Não envia o log, pois o último envio foi recente
		}
	}

	// Atualiza o timestamp do último envio para esse arquivo
	lastSent[filename] = now
	return true
}

// Obtém as últimas 'n' linhas de um arquivo
func getLastLines(filename string, n int) string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Erro ao abrir arquivo:", err)
		return ""
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}
	return strings.Join(lines, "\n")
}

// Obtém o nome da aplicação com base no diretório do arquivo
func getApplicationName(filePath string) string {
	parts := strings.Split(filePath, string(os.PathSeparator))
	if len(parts) > 2 {
		return parts[len(parts)-2] // Pega o nome da pasta anterior ao arquivo
	}
	return "Desconhecido"
}

func normalizeFilename(filename string) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	if ext == ".txt" {
		return filename
	}
	return base + ".txt"
}
