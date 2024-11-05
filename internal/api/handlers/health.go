package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/arkishshah/go-infra-provisioner/pkg/logger"
)

type HealthHandler struct {
	logger *logger.Logger
}

func NewHealthHandler(logger *logger.Logger) *HealthHandler {
	return &HealthHandler{logger: logger}
}

func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
