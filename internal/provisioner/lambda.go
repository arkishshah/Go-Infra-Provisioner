package provisioner

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// Helper function to create ZIP file bytes
func createZipBytes(functionCode string) []byte {
	// Create a buffer to write our zip to
	buf := new(bytes.Buffer)

	// Create a new zip archive
	w := zip.NewWriter(buf)

	// Create a new file in the zip
	f, err := w.Create("index.js")
	if err != nil {
		return nil
	}

	// Write the function code to the file
	_, err = io.WriteString(f, functionCode)
	if err != nil {
		return nil
	}

	// Make sure to close the zip writer
	err = w.Close()
	if err != nil {
		return nil
	}

	return buf.Bytes()
}

func (p *ResourceProvisioner) createLambdaFunction(ctx context.Context, functionName, roleARN, targetBucket string) (string, error) {
	p.logger.Info(fmt.Sprintf("Creating Lambda function: %s", functionName))

	// Lambda function code (basic log processor)
	functionCode := fmt.Sprintf(`
exports.handler = async (event) => {
    const AWS = require('aws-sdk');
    const s3 = new AWS.S3();
    
    try {
        // Process CloudWatch Logs
        const logData = event.detail.requestParameters;
        
        // Store in S3
        await s3.putObject({
            Bucket: '%s',
            Key: 'logs/' + new Date().toISOString() + '.json',
            Body: JSON.stringify(logData)
        }).promise();
        
        console.log('Successfully processed logs');
        return {
            statusCode: 200,
            body: 'Logs processed successfully'
        };
    } catch (error) {
        console.error('Error:', error);
        throw error;
    }
};`, targetBucket)

	// Create ZIP file containing the function code
	zipBytes := createZipBytes(functionCode)
	if zipBytes == nil {
		return "", fmt.Errorf("failed to create zip file for lambda function")
	}

	// Create Lambda function
	createResult, err := p.lambdaClient.CreateFunction(ctx, &lambda.CreateFunctionInput{
		FunctionName: aws.String(functionName),
		Role:         aws.String(roleARN),
		Handler:      aws.String("index.handler"),
		Code: &types.FunctionCode{
			ZipFile: zipBytes,
		},
		Runtime: types.RuntimeNodejs18x,
		Environment: &types.Environment{
			Variables: map[string]string{
				"TARGET_BUCKET": targetBucket,
			},
		},
		Timeout:    aws.Int32(30),
		MemorySize: aws.Int32(128),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create lambda function: %w", err)
	}

	return *createResult.FunctionArn, nil
}
func (p *ResourceProvisioner) deleteLambdaFunction(ctx context.Context, functionName string) error {
	p.logger.Info(fmt.Sprintf("Deleting Lambda function: %s", functionName))

	_, err := p.lambdaClient.DeleteFunction(ctx, &lambda.DeleteFunctionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete lambda function: %w", err)
	}

	return nil
}
