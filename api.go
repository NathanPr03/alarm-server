package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

type SoundResponse struct {
	Message  string `json:"message"`
	Filename string `json:"filename"`
}

type SoundListResponse struct {
	Sounds []string `json:"sounds"`
}

type ScheduleRequest struct {
	Timestamp     string `json:"timestamp"`
	SoundFileName string `json:"sound"`
}

type ScheduleResponse struct {
	Message string `json:"message"`
}

type Server struct {
	sync.Mutex
	alarmTimer *time.Timer
	lightTimer *time.Timer
	db         *sql.DB
	soundDir   string
	logger     *zap.Logger
}

func NewServer(logger *zap.Logger, db *sql.DB) (*Server, error) {
	soundDir := "sounds"
	if err := os.MkdirAll(soundDir, 0755); err != nil {
		return nil, err
	}

	return &Server{
		logger:   logger,
		db:       db,
		soundDir: soundDir,
	}, nil
}

func (server *Server) ScheduleHandler(w http.ResponseWriter, r *http.Request) {
	contextLogger := server.logger.With(zap.String("url", r.URL.String()))
	contextLogger = contextLogger.With(zap.String("method", r.Method))

	if r.Method != http.MethodPost {
		contextLogger.Warn("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		contextLogger.Warn("Invalid JSON")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	reqMarshalled, _ := json.Marshal(req)
	contextLogger = contextLogger.With(zap.String("body", string(reqMarshalled)))

	scheduleTime, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		http.Error(w, "Invalid timestamp format. Use RFC3339.", http.StatusBadRequest)
		return
	}

	alarmDuration := time.Until(scheduleTime)
	if alarmDuration <= 0 {
		http.Error(w, "Timestamp must be in the future", http.StatusBadRequest)
		return
	}

	lightTriggerTime := scheduleTime.Add(-15 * time.Minute)
	lightDuration := time.Until(lightTriggerTime)

	server.Lock()
	defer server.Unlock()

	// Cancel previous timers if they exist
	if server.lightTimer != nil {
		server.lightTimer.Stop()
	}
	if server.alarmTimer != nil {
		server.alarmTimer.Stop()
	}

	if lightDuration > 0 {
		server.lightTimer = time.AfterFunc(lightDuration, func() {
			WhenLightTriggered(contextLogger)
		})
	}

	server.alarmTimer = time.AfterFunc(alarmDuration, func() {
		WhenAlarmTriggered(contextLogger, req.SoundFileName)
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ScheduleResponse{Message: "Timer scheduled successfully"})
}

func (server *Server) UploadSoundHandler(w http.ResponseWriter, r *http.Request) {
	contextLogger := server.logger.With(zap.String("url", r.URL.String()))
	contextLogger = contextLogger.With(zap.String("filename", r.Header.Get("Content-Disposition")))
	contextLogger = contextLogger.With(zap.String("method", r.Method))

	if r.Method != http.MethodPost {
		contextLogger.Warn("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		contextLogger.Warn("Invalid file upload")
		http.Error(w, "Invalid file upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save file to sound directory
	soundPath := filepath.Join(server.soundDir, header.Filename)
	outFile, err := os.Create(soundPath)
	if err != nil {
		contextLogger.Error("Failed to save file", zap.Error(err))
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		contextLogger.Error("Failed to write file", zap.Error(err))
		http.Error(w, "Failed to write file", http.StatusInternalServerError)
		return
	}

	// Store metadata in SQLite
	_, err = server.db.Exec("INSERT INTO sounds (filename) VALUES (?)", header.Filename)
	if err != nil {
		contextLogger.Error("Failed to store sound metadata", zap.Error(err))
		http.Error(w, "Failed to store sound metadata", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SoundResponse{Message: "Sound uploaded successfully", Filename: header.Filename})
}

func (server *Server) ListSoundsHandler(w http.ResponseWriter, r *http.Request) {
	contextLogger := server.logger.With(zap.String("url", r.URL.String()))
	contextLogger = contextLogger.With(zap.String("method", r.Method))

	rows, err := server.db.Query("SELECT filename FROM sounds")
	if err != nil {
		contextLogger.Error("Failed to fetch sounds", zap.Error(err))
		http.Error(w, "Failed to fetch sounds", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var sounds []string
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			contextLogger.Error("Failed to read sound", zap.String("filename", filename), zap.Error(err))
			http.Error(w, "Failed to read sounds", http.StatusInternalServerError)
			return
		}
		sounds = append(sounds, filename)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SoundListResponse{Sounds: sounds})
}
