package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/imama2/Genzite-Backend/internal/core/broker"
	"github.com/imama2/Genzite-Backend/internal/core/config"
	"github.com/imama2/Genzite-Backend/internal/services/agents"
	"github.com/imama2/Genzite-Backend/internal/services/chatbot"
	"github.com/imama2/Genzite-Backend/internal/services/websocket"
)

var addr = flag.String("addr", ":8081", "http service address")

func main() {
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Initialize services
	agentService := agents.NewService()
	brokerService, err := broker.NewClient(cfg, logger)
	if err != nil {
		logger.Error("Failed to connect to broker", "error", err)
		os.Exit(1)
	}

	chatService := chatbot.NewService(logger, agentService, brokerService)
	hub := websocket.NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, chatService, w, r)
	})

	logger.Info("Starting orchestrator server", "addr", *addr)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
