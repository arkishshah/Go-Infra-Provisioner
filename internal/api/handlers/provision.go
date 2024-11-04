package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/arkishshah/go-infra-provisioner/internal/config"
	"github.com/arkishshah/go-infra-provisioner/internal/models"
	"github.com/arkishshah/go-infra-provisioner/internal/provisioner"
	"github.com/arkishshah/go-infra-provisioner/pkg/awsclient"
	"github.com/arkishshah/go-infra-provisioner/pkg/logger"
)

type ProvisionHandler struct {
	provisioner *provisioner.ResourceProvisioner
	logger      *logger.Logger
}

func NewProvisionHandler(cfg *config.Config, awsClient *awsclient.AWSClient, logger *logger.Logger) *ProvisionHandler {
	return &ProvisionHandler{
		provisioner: provisioner.NewResourceProvisioner(cfg, awsClient),
		logger:      logger,
	}
}

func (h *ProvisionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req models.ProvisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request:", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.provisioner.ProvisionClientResources(r.Context(), &req); err != nil {
		h.logger.Error("Failed to provision resources:", err)
		http.Error(w, "Failed to provision resources", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "provisioned"})
}
