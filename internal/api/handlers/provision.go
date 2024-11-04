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
	config      *config.Config
}

func NewProvisionHandler(cfg *config.Config, awsClient *awsclient.AWSClient, logger *logger.Logger) *ProvisionHandler {
	return &ProvisionHandler{
		provisioner: provisioner.NewResourceProvisioner(cfg, awsClient, logger),
		logger:      logger,
		config:      cfg,
	}
}

func (h *ProvisionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req models.ProvisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateRequest(&req); err != nil {
		h.logger.Error("Invalid request:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call provisioner
	response, err := h.provisioner.ProvisionClientResources(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to provision resources:", err)
		http.Error(w, "Failed to provision resources", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Successfully provisioned resources for client:", req.ClientID)
}

func (h *ProvisionHandler) validateRequest(req *models.ProvisionRequest) error {
	if req.ClientID == "" {
		return &models.ProvisionError{
			Code:    "INVALID_REQUEST",
			Message: "client_id is required",
		}
	}

	if req.ClientName == "" {
		return &models.ProvisionError{
			Code:    "INVALID_REQUEST",
			Message: "client_name is required",
		}
	}

	// Add any additional validation as needed
	return nil
}
