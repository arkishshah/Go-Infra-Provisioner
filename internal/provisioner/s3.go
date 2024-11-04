package provisioner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (p *ResourceProvisioner) createS3Bucket(ctx context.Context, bucketName string) error {
	p.logger.Info(fmt.Sprintf("Creating S3 bucket: %s", bucketName))

	_, err := p.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	// Enable versioning
	_, err = p.s3Client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusEnabled,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to enable bucket versioning: %w", err)
	}

	return nil
}

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

		p.logger.Info(fmt.Sprintf("Waiting for bucket to be available, attempt: %d", i+1))
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("bucket did not become available within the expected time")
}

func (p *ResourceProvisioner) deleteS3Bucket(ctx context.Context, bucketName string) error {
	p.logger.Info(fmt.Sprintf("Deleting S3 bucket: %s", bucketName))

	_, err := p.s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}
func (p *ResourceProvisioner) applyBucketPolicy(ctx context.Context, bucketName, roleARN string) error {
	p.logger.Info(fmt.Sprintf("Applying bucket policy for: %s with role ARN: %s", bucketName, roleARN))

	// Define the bucket policy with the provided role ARN as Principal
	bucketPolicy := fmt.Sprintf(`{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Sid": "AllowRoleAccess",
                "Effect": "Allow",
                "Principal": {
                    "AWS": "%s"
                },
                "Action": [
                    "s3:GetObject",
                    "s3:PutObject",
                    "s3:ListBucket",
                    "s3:DeleteObject"
                ],
                "Resource": [
                    "arn:aws:s3:::%s",
                    "arn:aws:s3:::%s/*"
                ]
            }
        ]
    }`, roleARN, bucketName, bucketName)

	p.logger.Info(fmt.Sprintf("Bucket policy to be applied: %s", bucketPolicy))

	// Retry mechanism
	const maxRetries = 5
	const retryDelay = 5 * time.Second

	for i := 1; i <= maxRetries; i++ {
		_, err := p.s3Client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
			Bucket: aws.String(bucketName),
			Policy: aws.String(bucketPolicy),
		})

		if err == nil {
			p.logger.Info("Successfully applied bucket policy")
			return nil // Success
		}

		// Log and retry if an error occurs
		p.logger.Error(fmt.Sprintf("Failed to apply bucket policy (attempt %d/%d): %v", i, maxRetries, err))

		// Only retry if it's an Invalid Principal error
		if i < maxRetries {
			p.logger.Info("Retrying to apply bucket policy after delay...")
			time.Sleep(retryDelay)
		} else {
			return fmt.Errorf("failed to apply bucket policy after %d attempts: %w", maxRetries, err)
		}
	}

	return nil
}
