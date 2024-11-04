package provisioner

import (
	"context"
	"fmt"
	"time"

	"github.com/arkishshah/go-infra-provisioner/internal/config"
	"github.com/arkishshah/go-infra-provisioner/internal/models"
	"github.com/arkishshah/go-infra-provisioner/pkg/awsclient"
	"github.com/arkishshah/go-infra-provisioner/pkg/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
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

// ProvisionClientResources handles the complete provisioning process for a client
func (p *ResourceProvisioner) ProvisionClientResources(ctx context.Context, req *models.ProvisionRequest) (*models.ProvisionResponse, error) {
	p.logger.Info("Starting resource provisioning for client:", req.ClientID)

	// Generate resource names
	bucketName := fmt.Sprintf("%s-%s-bucket", p.config.Environment, req.ClientID)
	roleName := fmt.Sprintf("%s-%s-role", p.config.Environment, req.ClientID)

	// Create S3 bucket
	if err := p.createS3Bucket(ctx, bucketName); err != nil {
		return nil, fmt.Errorf("failed to create S3 bucket: %w", err)
	}

	// Wait for bucket to be available
	if err := p.waitForBucketAvailability(ctx, bucketName); err != nil {
		p.logger.Error("Bucket availability check failed, cleaning up:", err)
		if cleanupErr := p.deleteS3Bucket(ctx, bucketName); cleanupErr != nil {
			return nil, fmt.Errorf("bucket creation failed and cleanup failed: %v, cleanup error: %v", err, cleanupErr)
		}
		return nil, err
	}

	// Create IAM role
	roleARN, err := p.createIAMRole(ctx, roleName, bucketName)
	if err != nil {
		p.logger.Error("IAM role creation failed, cleaning up bucket:", err)
		if cleanupErr := p.deleteS3Bucket(ctx, bucketName); cleanupErr != nil {
			return nil, fmt.Errorf("role creation failed and cleanup failed: %v, cleanup error: %v", err, cleanupErr)
		}
		return nil, fmt.Errorf("failed to create IAM role: %w", err)
	}

	// Apply bucket policy
	if err := p.applyBucketPolicy(ctx, bucketName, roleARN); err != nil {
		p.logger.Error("Bucket policy application failed, cleaning up resources:", err)
		if cleanupErr := p.cleanupResources(ctx, bucketName, roleName); cleanupErr != nil {
			return nil, fmt.Errorf("policy application failed and cleanup failed: %v, cleanup error: %v", err, cleanupErr)
		}
		return nil, fmt.Errorf("failed to apply bucket policy: %w", err)
	}

	p.logger.Info("Successfully provisioned resources for client:", req.ClientID)

	return &models.ProvisionResponse{
		Status:     "success",
		BucketName: bucketName,
		RoleARN:    roleARN,
	}, nil
}

// waitForBucketAvailability waits for the bucket to be available

func (p *ResourceProvisioner) waitForBucketAvailability(ctx context.Context, bucketName string) error {
	maxRetries := 10
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		_, err := p.s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err == nil {
			return nil
		}

		p.logger.Info("Waiting for bucket to be available, attempt:", i+1)
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("bucket did not become available within the expected time")
}

// cleanupResources handles cleanup of all provisioned resources
func (p *ResourceProvisioner) cleanupResources(ctx context.Context, bucketName, roleName string) error {
	var errors []string

	// Delete bucket first
	if err := p.deleteS3Bucket(ctx, bucketName); err != nil {
		errors = append(errors, fmt.Sprintf("failed to delete bucket: %v", err))
	}

	// Delete role policy
	if err := p.cleanupIAMRole(ctx, roleName); err != nil {
		errors = append(errors, fmt.Sprintf("failed to cleanup IAM role: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	p.logger.Info("Successfully cleaned up all resources")
	return nil
}
