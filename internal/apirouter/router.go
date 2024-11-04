package apirouter

import (
	"github.com/arkishshah/go-infra-provisioner/internal/api/handlers"
	"github.com/arkishshah/go-infra-provisioner/internal/api/middleware"
	"github.com/arkishshah/go-infra-provisioner/internal/config"
	"github.com/arkishshah/go-infra-provisioner/pkg/awsclient"
	"github.com/arkishshah/go-infra-provisioner/pkg/logger"
	"github.com/gorilla/mux"
)

func NewRouter(cfg *config.Config, awsClient *awsclient.AWSClient, logger *logger.Logger) *mux.Router {
	r := mux.NewRouter()

	// Initialize handlers
	provisionHandler := handlers.NewProvisionHandler(cfg, awsClient, logger)
	healthHandler := handlers.NewHealthHandler(logger)

	// Add middleware
	r.Use(middleware.Logging(logger))

	// Routes
	r.HandleFunc("/health", healthHandler.Handle).Methods("GET")
	r.HandleFunc("/api/v1/provision", provisionHandler.Handle).Methods("POST")

	return r
}
