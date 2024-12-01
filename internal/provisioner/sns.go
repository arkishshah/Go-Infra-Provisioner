package provisioner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
)

func (p *ResourceProvisioner) createSNSTopic(ctx context.Context, topicName string) (string, error) {
	p.logger.Info(fmt.Sprintf("Creating SNS topic: %s", topicName))

	// Create SNS topic with correct Tag type
	result, err := p.snsClient.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: aws.String(topicName),
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
		return "", fmt.Errorf("failed to create SNS topic: %w", err)
	}

	// Set up topic policy
	policy := fmt.Sprintf(`{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": {
                    "Service": "cloudwatch.amazonaws.com"
                },
                "Action": "sns:Publish",
                "Resource": "%s"
            }
        ]
    }`, *result.TopicArn)

	_, err = p.snsClient.SetTopicAttributes(ctx, &sns.SetTopicAttributesInput{
		TopicArn:       result.TopicArn,
		AttributeName:  aws.String("Policy"),
		AttributeValue: aws.String(policy),
	})
	if err != nil {
		return "", fmt.Errorf("failed to set topic policy: %w", err)
	}

	return *result.TopicArn, nil
}
func (p *ResourceProvisioner) deleteSNSTopic(ctx context.Context, topicARN string) error {
	p.logger.Info(fmt.Sprintf("Deleting SNS topic: %s", topicARN))

	_, err := p.snsClient.DeleteTopic(ctx, &sns.DeleteTopicInput{
		TopicArn: aws.String(topicARN),
	})
	if err != nil {
		return fmt.Errorf("failed to delete SNS topic: %w", err)
	}

	return nil
}
