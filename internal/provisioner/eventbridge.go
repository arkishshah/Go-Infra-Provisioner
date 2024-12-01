package provisioner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

func (p *ResourceProvisioner) createEventRule(ctx context.Context, ruleName, logGroupName, lambdaARN string) error {
	p.logger.Info(fmt.Sprintf("Creating EventBridge rule: %s", ruleName))

	// Create rule pattern to match log events
	pattern := fmt.Sprintf(`{
        "source": ["aws.logs"],
        "detail-type": ["AWS API Call via CloudTrail"],
        "detail": {
            "eventSource": ["logs.amazonaws.com"],
            "eventName": ["PutLogEvents"],
            "requestParameters": {
                "logGroupName": ["%s"]
            }
        }
    }`, logGroupName)

	// Create the rule
	_, err := p.eventBridgeClient.PutRule(ctx, &eventbridge.PutRuleInput{
		Name:         aws.String(ruleName),
		Description:  aws.String(fmt.Sprintf("Process logs from %s", logGroupName)),
		EventPattern: aws.String(pattern),
		State:        types.RuleStateEnabled, // Fixed: Using the correct type
	})
	if err != nil {
		return fmt.Errorf("failed to create event rule: %w", err)
	}

	// Add target (Lambda function)
	_, err = p.eventBridgeClient.PutTargets(ctx, &eventbridge.PutTargetsInput{
		Rule: aws.String(ruleName),
		Targets: []types.Target{
			{
				Id:  aws.String("ProcessLogsFunction"),
				Arn: aws.String(lambdaARN),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add target to event rule: %w", err)
	}

	return nil
}
func (p *ResourceProvisioner) deleteEventRule(ctx context.Context, ruleName string) error {
	p.logger.Info(fmt.Sprintf("Deleting EventBridge rule: %s", ruleName))

	_, err := p.eventBridgeClient.DeleteRule(ctx, &eventbridge.DeleteRuleInput{
		Name: aws.String(ruleName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete event rule: %w", err)
	}

	return nil
}
