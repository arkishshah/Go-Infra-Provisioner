package provisioner

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// createS3Bucket creates an S3 bucket with encryption and versioning
func (p *ResourceProvisioner) createS3Bucket(ctx context.Context, bucketName string) error {
	p.logger.Info("Creating S3 bucket:", bucketName)

	_, err := p.s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(p.config.AWSRegion),
		},
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

	// Enable encryption
	_, err = p.s3Client.PutBucketEncryption(ctx, &s3.PutBucketEncryptionInput{
		Bucket: aws.String(bucketName),
		ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
			Rules: []types.ServerSideEncryptionRule{
				{
					ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{
						SSEAlgorithm: types.ServerSideEncryptionAes256,
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to enable bucket encryption: %w", err)
	}

	return nil
}

// deleteS3Bucket deletes an S3 bucket
func (p *ResourceProvisioner) deleteS3Bucket(ctx context.Context, bucketName string) error {
	p.logger.Info("Deleting S3 bucket:", bucketName)

	_, err := p.s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	return err
}

// applyBucketPolicy applies the bucket policy allowing access from the IAM role
func (p *ResourceProvisioner) applyBucketPolicy(ctx context.Context, bucketName, roleARN string) error {
	p.logger.Info("Applying bucket policy for:", bucketName)

	bucketPolicy := fmt.Sprintf(`{
        "Version": "2012-10-17",
        "Statement": [{
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
        }]
    }`, roleARN, bucketName, bucketName)

	_, err := p.s3Client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucketName),
		Policy: aws.String(bucketPolicy),
	})
	return err
}
