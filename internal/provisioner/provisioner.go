package provisioner

import (
	"context"
	"fmt"

	"github.com/arkishshah/go-infra-provisioner/internal/config"
	"github.com/arkishshah/go-infra-provisioner/internal/models"
	"github.com/arkishshah/go-infra-provisioner/pkg/awsclient"
	"github.com/arkishshah/go-infra-provisioner/pkg/logger"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type ResourceProvisioner struct {
	s3Client             *s3.Client
	iamClient            *iam.Client
	cloudwatchClient     *cloudwatch.Client
	cloudwatchLogsClient *cloudwatchlogs.Client
	eventBridgeClient    *eventbridge.Client
	lambdaClient         *lambda.Client
	snsClient            *sns.Client
	config               *config.Config
	logger               *logger.Logger
}

func NewResourceProvisioner(cfg *config.Config, awsClient *awsclient.AWSClient, logger *logger.Logger) *ResourceProvisioner {
	return &ResourceProvisioner{
		s3Client:             awsClient.S3Client,
		iamClient:            awsClient.IAMClient,
		cloudwatchClient:     awsClient.CloudWatchClient,
		cloudwatchLogsClient: awsClient.CloudWatchLogsClient,
		eventBridgeClient:    awsClient.EventBridgeClient,
		lambdaClient:         awsClient.LambdaClient,
		snsClient:            awsClient.SNSClient,
		config:               cfg,
		logger:               logger,
	}
}

func (p *ResourceProvisioner) ProvisionClientResources(ctx context.Context, req *models.ProvisionRequest) (*models.ProvisionResponse, error) {
	p.logger.Info(fmt.Sprintf("Starting resource provisioning for client: %s", req.ClientID))

	// Generate resource names
	bucketName := fmt.Sprintf("%s-%s-bucket", p.config.Environment, req.ClientID)
	roleName := fmt.Sprintf("%s-%s-role", p.config.Environment, req.ClientID)
	logGroupName := fmt.Sprintf("/aws/client/%s/%s", p.config.Environment, req.ClientID)
	ruleName := fmt.Sprintf("%s-%s-rule", p.config.Environment, req.ClientID)
	lambdaName := fmt.Sprintf("%s-%s-processor", p.config.Environment, req.ClientID)
	topicName := fmt.Sprintf("%s-%s-alerts", p.config.Environment, req.ClientID)

	// Create S3 bucket
	err := p.createS3Bucket(ctx, bucketName)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to create S3 bucket: %v", err))
		return nil, fmt.Errorf("failed to create S3 bucket: %w", err)
	}

	// Create IAM role
	roleARN, err := p.createIAMRole(ctx, roleName, bucketName, logGroupName, lambdaName)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to create IAM role: %v", err))
		p.deleteS3Bucket(ctx, bucketName)
		return nil, fmt.Errorf("failed to create IAM role: %w", err)
	}

	// Create CloudWatch Log Group
	err = p.createLogGroup(ctx, logGroupName)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to create log group: %v", err))
		p.cleanup(ctx, &cleanupConfig{
			bucketName: bucketName,
			roleName:   roleName,
		})
		return nil, fmt.Errorf("failed to create log group: %w", err)
	}

	// Create Lambda Function
	lambdaARN, err := p.createLambdaFunction(ctx, lambdaName, roleARN, bucketName)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to create lambda function: %v", err))
		p.cleanup(ctx, &cleanupConfig{
			bucketName:   bucketName,
			roleName:     roleName,
			logGroupName: logGroupName,
		})
		return nil, fmt.Errorf("failed to create lambda function: %w", err)
	}

	// Create EventBridge Rule
	err = p.createEventRule(ctx, ruleName, logGroupName, lambdaARN)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to create event rule: %v", err))
		p.cleanup(ctx, &cleanupConfig{
			bucketName:   bucketName,
			roleName:     roleName,
			logGroupName: logGroupName,
			lambdaName:   lambdaName,
		})
		return nil, fmt.Errorf("failed to create event rule: %w", err)
	}

	// Create SNS Topic
	topicARN, err := p.createSNSTopic(ctx, topicName)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to create SNS topic: %v", err))
		p.cleanup(ctx, &cleanupConfig{
			bucketName:   bucketName,
			roleName:     roleName,
			logGroupName: logGroupName,
			lambdaName:   lambdaName,
			ruleName:     ruleName,
		})
		return nil, fmt.Errorf("failed to create SNS topic: %w", err)
	}

	// Set up CloudWatch Alarms
	err = p.setupCloudWatchAlarms(ctx, logGroupName, topicARN, req.ClientID)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to set up alarms: %v", err))
		p.cleanup(ctx, &cleanupConfig{
			bucketName:   bucketName,
			roleName:     roleName,
			logGroupName: logGroupName,
			lambdaName:   lambdaName,
			ruleName:     ruleName,
			topicARN:     topicARN,
		})
		return nil, fmt.Errorf("failed to set up alarms: %w", err)
	}

	response := &models.ProvisionResponse{
		Status:       "success",
		BucketName:   bucketName,
		RoleARN:      roleARN,
		LogGroupName: logGroupName,
		LambdaARN:    lambdaARN,
		TopicARN:     topicARN,
	}

	p.logger.Info("Successfully provisioned all resources")
	return response, nil
}

// Add this struct and method in your provisioner.go file, after the ProvisionClientResources function

type cleanupConfig struct {
	bucketName   string
	roleName     string
	logGroupName string
	lambdaName   string
	ruleName     string
	topicARN     string
}

func (p *ResourceProvisioner) cleanup(ctx context.Context, config *cleanupConfig) {
	if config.bucketName != "" {
		if err := p.deleteS3Bucket(ctx, config.bucketName); err != nil {
			p.logger.Error(fmt.Sprintf("Failed to cleanup S3 bucket: %v", err))
		}
	}

	if config.roleName != "" {
		if err := p.cleanupIAMRole(ctx, config.roleName); err != nil {
			p.logger.Error(fmt.Sprintf("Failed to cleanup IAM role: %v", err))
		}
	}

	if config.logGroupName != "" {
		if err := p.deleteLogGroup(ctx, config.logGroupName); err != nil {
			p.logger.Error(fmt.Sprintf("Failed to cleanup log group: %v", err))
		}
	}

	if config.lambdaName != "" {
		if err := p.deleteLambdaFunction(ctx, config.lambdaName); err != nil {
			p.logger.Error(fmt.Sprintf("Failed to cleanup lambda function: %v", err))
		}
	}

	if config.ruleName != "" {
		if err := p.deleteEventRule(ctx, config.ruleName); err != nil {
			p.logger.Error(fmt.Sprintf("Failed to cleanup event rule: %v", err))
		}
	}

	if config.topicARN != "" {
		if err := p.deleteSNSTopic(ctx, config.topicARN); err != nil {
			p.logger.Error(fmt.Sprintf("Failed to cleanup SNS topic: %v", err))
		}
	}
}
