package provisioner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

func (p *ResourceProvisioner) setupCloudWatchAlarms(ctx context.Context, logGroupName, snsTopicArn, clientID string) error {
	p.logger.Info("Setting up CloudWatch Alarms")

	// Error Rate Alarm
	_, err := p.cloudwatchClient.PutMetricAlarm(ctx, &cloudwatch.PutMetricAlarmInput{
		AlarmName:          aws.String(fmt.Sprintf("%s-error-rate-alarm", clientID)),
		AlarmDescription:   aws.String("Alert when error rate exceeds threshold"),
		MetricName:         aws.String("ErrorCount"),
		Namespace:          aws.String("Custom/ClientLogs"),
		Statistic:          types.StatisticSum, // Fixed this
		Period:             aws.Int32(300),
		EvaluationPeriods:  aws.Int32(1),
		Threshold:          aws.Float64(10),
		ComparisonOperator: types.ComparisonOperatorGreaterThanThreshold,
		AlarmActions:       []string{snsTopicArn},
		Dimensions: []types.Dimension{
			{
				Name:  aws.String("ClientID"),
				Value: aws.String(clientID),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create error rate alarm: %w", err)
	}

	// Log Volume Alarm
	_, err = p.cloudwatchClient.PutMetricAlarm(ctx, &cloudwatch.PutMetricAlarmInput{
		AlarmName:          aws.String(fmt.Sprintf("%s-log-volume-alarm", clientID)),
		AlarmDescription:   aws.String("Alert on unusual log volume"),
		MetricName:         aws.String("IncomingLogEvents"),
		Namespace:          aws.String("AWS/Logs"),
		Statistic:          types.StatisticSum, // Fixed this
		Period:             aws.Int32(300),
		EvaluationPeriods:  aws.Int32(2),
		Threshold:          aws.Float64(1000),
		ComparisonOperator: types.ComparisonOperatorGreaterThanThreshold,
		AlarmActions:       []string{snsTopicArn},
		Dimensions: []types.Dimension{
			{
				Name:  aws.String("LogGroupName"),
				Value: aws.String(logGroupName),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create log volume alarm: %w", err)
	}

	return nil
}
func (p *ResourceProvisioner) createLogGroup(ctx context.Context, logGroupName string) error {
	p.logger.Info(fmt.Sprintf("Creating CloudWatch Log Group: %s", logGroupName))

	_, err := p.cloudwatchLogsClient.CreateLogGroup(ctx, &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	})
	if err != nil {
		return fmt.Errorf("failed to create log group: %w", err)
	}

	return nil
}

func (p *ResourceProvisioner) deleteLogGroup(ctx context.Context, logGroupName string) error {
	p.logger.Info(fmt.Sprintf("Deleting CloudWatch Log Group: %s", logGroupName))

	_, err := p.cloudwatchLogsClient.DeleteLogGroup(ctx, &cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(logGroupName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete log group: %w", err)
	}

	return nil
}
