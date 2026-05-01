package handler

import (
	"encoding/json"
	"go-banking-service/internal/logger"
	"go-banking-service/internal/middleware"
	"go-banking-service/internal/service"
	"net/http"
	"strconv"
)

type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
}

func NewAnalyticsHandler(analyticsService service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

func (h *AnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		d, err := strconv.Atoi(daysStr)
		if err == nil && d > 0 {
			days = d
		}
	}

	resp, err := h.analyticsService.GetAnalytics(r.Context(), userID, days)
	if err != nil {
		logger.Error("Failed to get analytics via handler", "error", err, "user_id", userID)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
