package provisioner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func (p *ResourceProvisioner) createS3Bucket(ctx context.Context, bucketName string) error {
	p.logger.Info(fmt.Sprintf("Creating S3 bucket: %s", bucketName))

	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	if p.config.AWSRegion != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(p.config.AWSRegion),
		}
	}

	_, err := p.s3Client.CreateBucket(ctx, input)
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
		p.logger.Error(fmt.Sprintf("Failed to enable versioning: %v", err))
	}

	// Setup lifecycle rules for log archival
	lifecycleRule := &types.LifecycleRule{
		Status: types.ExpirationStatusEnabled,
		Transitions: []types.Transition{
			{
				Days:         aws.Int32(30),
				StorageClass: types.TransitionStorageClassStandardIa,
			},
			{
				Days:         aws.Int32(90),
				StorageClass: types.TransitionStorageClassGlacier,
			},
		},
	}

	_, err = p.s3Client.PutBucketLifecycleConfiguration(ctx, &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucketName),
		LifecycleConfiguration: &types.BucketLifecycleConfiguration{
			Rules: []types.LifecycleRule{*lifecycleRule},
		},
	})
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to set lifecycle rules: %v", err))
	}

	return nil
}
func (p *ResourceProvisioner) deleteS3Bucket(ctx context.Context, bucketName string) error {
	p.logger.Info(fmt.Sprintf("Deleting S3 bucket: %s", bucketName))

	_, err := p.s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}
	return nil
}
