package main

import (
	"go.uber.org/zap"
	"net/http"
)

func main() {
	logger, _ := zap.NewProduction(
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	server, err := NewServer(logger)
	if err != nil {
		logger.Error("Failed to create server", zap.Error(err))
	}

	defer logger.Sync()
	mux := http.NewServeMux()
	mux.HandleFunc("/schedule", server.ScheduleHandler)

	logger.Info("Started server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		logger.Error("Server creating failed", zap.Error(err))
	}
}
