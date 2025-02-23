package main

import (
	"alarm-server/migrations"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	logger, _ := zap.NewProduction(
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	db, err := migrations.InitDatabase(migrations.DbPath)
	if err != nil {
		logger.Error("Failed to initialize database", zap.Error(err))
		return
	}

	defer db.Close()

	server, err := NewServer(logger, db)
	if err != nil {
		logger.Error("Failed to create server", zap.Error(err))
	}

	defer logger.Sync()
	mux := http.NewServeMux()
	mux.HandleFunc("/schedule", server.ScheduleHandler)
	mux.HandleFunc("/sounds", server.ListSoundsHandler)
	mux.HandleFunc("/sound", server.UploadSoundHandler)

	logger.Info("Started server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		logger.Error("Server creating failed", zap.Error(err))
	}
}
