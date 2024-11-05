package awsclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AWSClient struct {
	S3Client  *s3.Client
	IAMClient *iam.Client
}

func NewAWSClient(ctx context.Context) (*AWSClient, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Initialize clients
	return &AWSClient{
		S3Client:  s3.NewFromConfig(cfg),
		IAMClient: iam.NewFromConfig(cfg),
	}, nil
}
