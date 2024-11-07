#!/bin/bash

echo "🧹 Cleaning up resources..."

# Get client ID from argument or use default
CLIENT_ID=${1:-"test-client-001"}
ENVIRONMENT=${ENVIRONMENT:-"dev"}

# Construct resource names
BUCKET_NAME="${ENVIRONMENT}-${CLIENT_ID}-bucket"
ROLE_NAME="${ENVIRONMENT}-${CLIENT_ID}-role"

# Delete S3 bucket
echo "Deleting S3 bucket: $BUCKET_NAME"
aws s3 rb "s3://$BUCKET_NAME" --force

# Delete IAM role
echo "Deleting IAM role: $ROLE_NAME"
aws iam delete-role-policy --role-name "$ROLE_NAME" --policy-name "${ROLE_NAME}-policy"
aws iam delete-role --role-name "$ROLE_NAME"

echo "Cleanup complete!"
