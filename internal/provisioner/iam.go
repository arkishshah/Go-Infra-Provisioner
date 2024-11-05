package provisioner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

func (p *ResourceProvisioner) createIAMRole(ctx context.Context, roleName, bucketName string) (string, error) {
	p.logger.Info(fmt.Sprintf("Creating IAM role: %s", roleName))

	// Define the trust policy to allow EC2 instances to assume the role
	assumeRolePolicy := `{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": {
                    "Service": "ec2.amazonaws.com"
                },
                "Action": "sts:AssumeRole"
            }
        ]
    }`

	// Create the IAM role with the trust policy
	roleResult, err := p.iamClient.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(assumeRolePolicy),
		Description:              aws.String(fmt.Sprintf("Role for client bucket access: %s", bucketName)),
		Tags: []types.Tag{
			{
				Key:   aws.String("Environment"),
				Value: aws.String(p.config.Environment),
			},
			{
				Key:   aws.String("ManagedBy"),
				Value: aws.String("Provisioner"),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create role: %w", err)
	}

	// Use an explicit retry mechanism to wait for the IAM role to be fully propagated
	if err := p.waitForRolePropagation(ctx, roleName); err != nil {
		return "", fmt.Errorf("IAM role propagation failed: %w", err)
	}

	// Attach an inline policy for the IAM role to access the S3 bucket
	rolePolicy := fmt.Sprintf(`{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
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
    }`, bucketName, bucketName)

	_, err = p.iamClient.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyName:     aws.String(fmt.Sprintf("%s-policy", roleName)),
		PolicyDocument: aws.String(rolePolicy),
	})
	if err != nil {
		return "", fmt.Errorf("failed to attach inline policy to role: %w", err)
	}

	return *roleResult.Role.Arn, nil
}

// Enhanced role propagation check with retries
func (p *ResourceProvisioner) waitForRolePropagation(ctx context.Context, roleName string) error {
	const maxRetries = 10
	const retryInterval = 5 * time.Second

	for i := 1; i <= maxRetries; i++ {
		_, err := p.iamClient.GetRole(ctx, &iam.GetRoleInput{RoleName: aws.String(roleName)})
		if err == nil {
			p.logger.Info("IAM role is now available.")
			return nil // Role is available
		}

		p.logger.Info(fmt.Sprintf("Waiting for IAM role propagation (attempt %d/%d)...", i, maxRetries))
		time.Sleep(retryInterval)
	}

	return fmt.Errorf("IAM role propagation timed out after %d retries", maxRetries)
}

func (p *ResourceProvisioner) deleteIAMRole(ctx context.Context, roleName string) error {
	// Delete role policy
	_, err := p.iamClient.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
		RoleName:   aws.String(roleName),
		PolicyName: aws.String(fmt.Sprintf("%s-policy", roleName)),
	})
	if err != nil {
		return fmt.Errorf("failed to delete role policy: %w", err)
	}

	// Delete role
	_, err = p.iamClient.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}
