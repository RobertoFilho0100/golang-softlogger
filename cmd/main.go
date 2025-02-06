package main

import (
	"fmt"
	"go-softlogger/config"
	"go-softlogger/internal/watcher"
	"time"
)

func main() {
	fmt.Println("Iniciando o monitor de logs...")
	go watcher.WatchLogs(config.LogDirs)

	// Mantém a aplicação rodando
	for {
		time.Sleep(time.Hour)
	}
}
