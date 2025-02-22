package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type ScheduleRequest struct {
	Timestamp string `json:"timestamp"`
}

type ScheduleResponse struct {
	Message string `json:"message"`
}

type Server struct {
	sync.Mutex
	alarmTimer *time.Timer
	lightTimer *time.Timer
	logger     *zap.Logger
}

func (server *Server) ScheduleHandler(w http.ResponseWriter, r *http.Request) {
	contextLogger := server.logger.With(zap.Any("method", r.Method))
	contextLogger = contextLogger.With(zap.String("url", r.URL.String()))
	contextLogger = contextLogger.With(zap.String("remoteAddress", r.RemoteAddr))

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

	contextLogger = contextLogger.With(zap.Any("requestBody", req))

	scheduleTime, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		contextLogger.Warn("Invalid timestamp format")
		http.Error(w, "Invalid timestamp format. Use RFC3339.", http.StatusBadRequest)
		return
	}

	alarmDuration := time.Until(scheduleTime)
	if alarmDuration <= 0 {
		contextLogger.Warn("Timestamp must be in the future")
		http.Error(w, "Timestamp must be in the future", http.StatusBadRequest)
		return
	}

	lightTriggerTime := scheduleTime.Add(-15 * time.Minute)
	lightDuration := time.Until(lightTriggerTime)

	server.Lock()
	// Cancel previous timers if they exist
	if server.lightTimer != nil {
		contextLogger.Info("Cancelling previous light timer")
		server.lightTimer.Stop()
	}
	if server.alarmTimer != nil {
		contextLogger.Info("Cancelling previous alarm timer")
		server.alarmTimer.Stop()
	}

	if lightDuration > 0 {
		server.lightTimer = time.AfterFunc(lightDuration, func() {
			WhenLightTriggered(contextLogger)
			contextLogger.Info("Light triggered", zap.String("timestamp", lightTriggerTime.Format(time.RFC3339)))
		})
	} else {
		contextLogger.Warn("Light trigger time is in the past", zap.Int("lightDuration", int(lightDuration.Seconds())))
	}

	server.alarmTimer = time.AfterFunc(alarmDuration, func() {
		WhenAlarmTriggered(contextLogger)
		contextLogger.Info("Alarm triggered", zap.String("timestamp", scheduleTime.Format(time.RFC3339)))
	})

	server.Unlock()

	contextLogger.Info("Timer scheduled successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ScheduleResponse{Message: "Timer scheduled successfully"})
}
