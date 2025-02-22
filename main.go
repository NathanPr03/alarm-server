package main

import (
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	logger, _ := zap.NewProduction(
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	server := &Server{
		logger: logger,
	}

	defer logger.Sync()
	mux := http.NewServeMux()
	mux.HandleFunc("/schedule", server.ScheduleHandler)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
