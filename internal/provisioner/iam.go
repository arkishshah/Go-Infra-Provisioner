package provisioner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

// createIAMRole creates an IAM role with necessary permissions
func (p *ResourceProvisioner) createIAMRole(ctx context.Context, roleName, bucketName string) (string, error) {
	p.logger.Info("Creating IAM role:", roleName)

	assumeRolePolicy := `{
        "Version": "2012-10-17",
        "Statement": [{
            "Effect": "Allow",
            "Principal": {
                "Service": "ec2.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }]
    }`

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
				Value: aws.String("InfraProvisioner"),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create role: %w", err)
	}

	// Attach policy
	bucketPolicy := fmt.Sprintf(`{
        "Version": "2012-10-17",
        "Statement": [{
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
        }]
    }`, bucketName, bucketName)

	_, err = p.iamClient.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyName:     aws.String(fmt.Sprintf("%s-policy", roleName)),
		PolicyDocument: aws.String(bucketPolicy),
	})
	if err != nil {
		return "", fmt.Errorf("failed to attach policy: %w", err)
	}

	return *roleResult.Role.Arn, nil
}

// cleanupIAMRole handles cleanup of IAM resources
func (p *ResourceProvisioner) cleanupIAMRole(ctx context.Context, roleName string) error {
	// First delete the role policy
	policyName := fmt.Sprintf("%s-policy", roleName)
	_, err := p.iamClient.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
		RoleName:   aws.String(roleName),
		PolicyName: aws.String(policyName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete role policy: %w", err)
	}

	// Then delete the role itself
	_, err = p.iamClient.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	p.logger.Info("Successfully cleaned up IAM role:", roleName)
	return nil
}
