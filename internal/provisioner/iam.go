package provisioner

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

func (p *ResourceProvisioner) createIAMRole(ctx context.Context, roleName, bucketName, logGroupName, lambdaName string) (string, error) {
	p.logger.Info(fmt.Sprintf("Creating IAM role: %s", roleName))

	// Updated trust relationship to properly allow Lambda
	assumeRolePolicy := `{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": {
                    "Service": "lambda.amazonaws.com"
                },
                "Action": "sts:AssumeRole"
            }
        ]
    }`

	// Create the role
	roleResult, err := p.iamClient.CreateRole(ctx, &iam.CreateRoleInput{
		RoleName:                 aws.String(roleName),
		AssumeRolePolicyDocument: aws.String(assumeRolePolicy),
		Description:              aws.String(fmt.Sprintf("Role for client: %s", bucketName)),
		Tags: []types.Tag{
			{Key: aws.String("Environment"), Value: aws.String(p.config.Environment)},
			{Key: aws.String("ManagedBy"), Value: aws.String("Provisioner")},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create role: %w", err)
	}

	// Add small delay to allow role to propagate
	time.Sleep(10 * time.Second)

	// Attach the existing policy
	policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/go-infra-policy", p.config.AWSAccountID)
	_, err = p.iamClient.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyArn),
	})
	if err != nil {
		return "", fmt.Errorf("failed to attach main policy: %w", err)
	}

	// Also attach AWS Lambda basic execution role
	_, err = p.iamClient.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to attach Lambda execution policy: %w", err)
	}

	// Add client-specific inline policy
	inlinePolicy := fmt.Sprintf(`{
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
            },
            {
                "Effect": "Allow",
                "Action": [
                    "logs:CreateLogStream",
                    "logs:PutLogEvents",
                    "logs:GetLogEvents",
                    "logs:FilterLogEvents"
                ],
                "Resource": [
                    "arn:aws:logs:%s:%s:log-group:%s:*"
                ]
            }
        ]
    }`, bucketName, bucketName, p.config.AWSRegion, p.config.AWSAccountID, logGroupName)

	_, err = p.iamClient.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyName:     aws.String(fmt.Sprintf("%s-policy", roleName)),
		PolicyDocument: aws.String(inlinePolicy),
	})
	if err != nil {
		return "", fmt.Errorf("failed to attach inline policy: %w", err)
	}

	return *roleResult.Role.Arn, nil
}

// Update cleanup to detach both policies
func (p *ResourceProvisioner) cleanupIAMRole(ctx context.Context, roleName string) error {
	p.logger.Info(fmt.Sprintf("Cleaning up IAM role: %s", roleName))

	// Detach the main policy
	policyArn := fmt.Sprintf("arn:aws:iam::%s:policy/go-infra-policy", p.config.AWSAccountID)
	_, err := p.iamClient.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyArn),
	})
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to detach main policy: %v", err))
	}

	// Detach Lambda execution policy
	_, err = p.iamClient.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
	})
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to detach Lambda execution policy: %v", err))
	}

	// Delete the inline policy
	_, err = p.iamClient.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
		RoleName:   aws.String(roleName),
		PolicyName: aws.String(fmt.Sprintf("%s-policy", roleName)),
	})
	if err != nil {
		return fmt.Errorf("failed to delete role policy: %w", err)
	}

	// Delete the role
	_, err = p.iamClient.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	p.logger.Info(fmt.Sprintf("Successfully cleaned up IAM role: %s", roleName))
	return nil
}
