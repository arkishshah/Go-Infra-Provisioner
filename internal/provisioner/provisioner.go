package provisioner

import (
	"context"
	"fmt"

	"github.com/arkishshah/go-infra-provisioner/internal/api/models"
	"github.com/arkishshah/go-infra-provisioner/internal/config"
	"github.com/arkishshah/go-infra-provisioner/pkg/awsclient"
)

type ResourceProvisioner struct {
	s3Client  *awsclient.S3Client
	iamClient *awsclient.IAMClient
	config    *config.Config
}

func NewResourceProvisioner(cfg *config.Config, awsClient *awsclient.AWSClient) *ResourceProvisioner {
	return &ResourceProvisioner{
		s3Client:  awsClient.S3Client,
		iamClient: awsClient.IAMClient,
		config:    cfg,
	}
}

func (p *ResourceProvisioner) ProvisionClientResources(ctx context.Context, req *models.ProvisionRequest) error {
	// Create S3 bucket
	bucketName := fmt.Sprintf("client-%s-bucket", req.ClientID)
	if err := p.createS3Bucket(ctx, bucketName); err != nil {
		return fmt.Errorf("failed to create S3 bucket: %w", err)
	}

	// Create IAM role
	roleName := fmt.Sprintf("client-%s-role", req.ClientID)
	if err := p.createIAMRole(ctx, roleName, bucketName); err != nil {
		// Cleanup bucket if role creation fails
		if cleanupErr := p.deleteS3Bucket(ctx, bucketName); cleanupErr != nil {
			return fmt.Errorf("role creation failed and cleanup failed: %v, cleanup error: %v", err, cleanupErr)
		}
		return fmt.Errorf("failed to create IAM role: %w", err)
	}

	return nil
}
