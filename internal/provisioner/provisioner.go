package provisioner

import (
	"context"
	"fmt"

	"github.com/arkishshah/go-infra-provisioner/internal/config"
	"github.com/arkishshah/go-infra-provisioner/internal/models"
	"github.com/arkishshah/go-infra-provisioner/pkg/awsclient"
	"github.com/arkishshah/go-infra-provisioner/pkg/logger"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ResourceProvisioner struct {
	s3Client  *s3.Client
	iamClient *iam.Client
	config    *config.Config
	logger    *logger.Logger
}

func NewResourceProvisioner(cfg *config.Config, awsClient *awsclient.AWSClient, logger *logger.Logger) *ResourceProvisioner {
	return &ResourceProvisioner{
		s3Client:  awsClient.S3Client,
		iamClient: awsClient.IAMClient,
		config:    cfg,
		logger:    logger,
	}
}

func (p *ResourceProvisioner) ProvisionClientResources(ctx context.Context, req *models.ProvisionRequest) (*models.ProvisionResponse, error) {
	p.logger.Info(fmt.Sprintf("Starting resource provisioning for client: %s", req.ClientID))

	// Generate resource names
	bucketName := fmt.Sprintf("%s-%s-bucket", p.config.Environment, req.ClientID)
	roleName := fmt.Sprintf("%s-%s-role", p.config.Environment, req.ClientID)

	p.logger.Info(fmt.Sprintf("Generated names - Bucket: %s, Role: %s", bucketName, roleName))

	// Create S3 bucket
	p.logger.Info("Creating S3 bucket...")
	if err := p.createS3Bucket(ctx, bucketName); err != nil {
		p.logger.Error(fmt.Sprintf("Failed to create S3 bucket: %v", err))
		return nil, fmt.Errorf("failed to create S3 bucket: %w", err)
	}

	// Wait for bucket availability
	p.logger.Info("Waiting for bucket availability...")
	if err := p.waitForBucketAvailability(ctx, bucketName); err != nil {
		p.logger.Error(fmt.Sprintf("Bucket availability check failed: %v", err))
		if cleanupErr := p.deleteS3Bucket(ctx, bucketName); cleanupErr != nil {
			p.logger.Error(fmt.Sprintf("Cleanup after failure also failed: %v", cleanupErr))
		}
		return nil, err
	}

	// Create IAM role
	p.logger.Info("Creating IAM role...")
	roleARN, err := p.createIAMRole(ctx, roleName, bucketName)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to create IAM role: %v", err))
		if cleanupErr := p.deleteS3Bucket(ctx, bucketName); cleanupErr != nil {
			p.logger.Error(fmt.Sprintf("Cleanup after role creation failure failed: %v", cleanupErr))
		}
		return nil, fmt.Errorf("failed to create IAM role: %w", err)
	}

	// Apply bucket policy
	p.logger.Info("Applying bucket policy...")
	if err := p.applyBucketPolicy(ctx, bucketName, roleARN); err != nil {
		p.logger.Error(fmt.Sprintf("Failed to apply bucket policy: %v", err))
		if cleanupErr := p.cleanupResources(ctx, bucketName, roleName); cleanupErr != nil {
			p.logger.Error(fmt.Sprintf("Cleanup after policy application failure failed: %v", cleanupErr))
		}
		return nil, fmt.Errorf("failed to apply bucket policy: %w", err)
	}

	p.logger.Info(fmt.Sprintf("Successfully provisioned resources for client: %s", req.ClientID))

	response := &models.ProvisionResponse{
		Status:     "success",
		BucketName: bucketName,
		RoleARN:    roleARN,
	}

	p.logger.Info(fmt.Sprintf("Returning response: %+v", response))
	return response, nil
}

func (p *ResourceProvisioner) cleanupResources(ctx context.Context, bucketName, roleName string) error {
	var errors []error

	// Delete bucket
	if err := p.deleteS3Bucket(ctx, bucketName); err != nil {
		errors = append(errors, fmt.Errorf("failed to delete bucket: %w", err))
	}

	// Delete role policy and role
	if err := p.deleteIAMRole(ctx, roleName); err != nil {
		errors = append(errors, fmt.Errorf("failed to delete role resources: %w", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}
	return nil
}
