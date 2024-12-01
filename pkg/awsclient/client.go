package awsclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs" // Fixed this import
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type AWSClient struct {
	S3Client             *s3.Client
	IAMClient            *iam.Client
	CloudWatchClient     *cloudwatch.Client
	CloudWatchLogsClient *cloudwatchlogs.Client
	EventBridgeClient    *eventbridge.Client
	LambdaClient         *lambda.Client
	SNSClient            *sns.Client
}

func NewAWSClient(ctx context.Context) (*AWSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &AWSClient{
		S3Client:             s3.NewFromConfig(cfg),
		IAMClient:            iam.NewFromConfig(cfg),
		CloudWatchClient:     cloudwatch.NewFromConfig(cfg),
		CloudWatchLogsClient: cloudwatchlogs.NewFromConfig(cfg),
		EventBridgeClient:    eventbridge.NewFromConfig(cfg),
		LambdaClient:         lambda.NewFromConfig(cfg),
		SNSClient:            sns.NewFromConfig(cfg),
	}, nil
}
