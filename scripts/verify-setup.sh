#!/bin/bash

echo "🔍 Starting verification..."

# Function to load .env file
load_env() {
    if [ -f .env ]; then
        echo "Loading .env file..."
        set -a  # automatically export all variables
        source .env
        set +a
        echo "Environment variables loaded successfully"
    else
        echo "❌ .env file not found in current directory"
        exit 1
    fi
}

# Load environment variables first
load_env

# Print loaded AWS configurations (masking sensitive data)
echo -e "\nLoaded AWS Configurations:"
echo "AWS_ACCESS_KEY_ID: ${AWS_ACCESS_KEY_ID:0:5}..."
echo "AWS_SECRET_ACCESS_KEY: ${AWS_SECRET_ACCESS_KEY:0:5}..."
echo "AWS_REGION: $AWS_REGION"
echo "AWS_ACCOUNT_ID: $AWS_ACCOUNT_ID"
echo "ENVIRONMENT: $ENVIRONMENT"

# Verify required variables are set
echo -e "\nVerifying required variables..."
required_vars=(
    "AWS_ACCESS_KEY_ID"
    "AWS_SECRET_ACCESS_KEY"
    "AWS_REGION"
    "AWS_ACCOUNT_ID"
    "ENVIRONMENT"
)

# Check all required variables
echo "Checking required environment variables..."
for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Missing required environment variable: $var"
        exit 1
    else
        echo "✅ $var is set"
    fi
done

# Verify AWS Configuration
echo "Verifying AWS Configuration..."
aws sts get-caller-identity
if [ $? -eq 0 ]; then
    echo "✅ AWS credentials are valid"
else
    echo "❌ AWS credentials are invalid"
    exit 1
fi

# Verify Service Role exists
echo "Verifying Service Role..."
aws iam get-role --role-name $(echo $SERVICE_ROLE_ARN | cut -d'/' -f2) 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✅ Service Role exists and is accessible"
else
    echo "❌ Service Role not found or not accessible"
    echo "Please ensure you've run terraform and updated .env with the correct SERVICE_ROLE_ARN"
    exit 1
fi

# First, ensure we have the KMS_KEY_ID
echo "Checking KMS_KEY_ID..."
if [ -z "$KMS_KEY_ID" ]; then
    echo "❌ KMS_KEY_ID is not set in .env"
    exit 1
fi
echo "KMS_KEY_ID is set to: $KMS_KEY_ID"

# Verify KMS Key
echo "Verifying KMS Key..."
# Try with different key ID formats
if aws kms describe-key --key-id "$KMS_KEY_ID" 2>/dev/null; then
    echo "✅ KMS Key exists and is accessible"
elif aws kms describe-key --key-id "arn:aws:kms:${AWS_REGION}:${AWS_ACCOUNT_ID}:key/${KMS_KEY_ID}" 2>/dev/null; then
    echo "✅ KMS Key exists and is accessible (using full ARN)"
else
    echo "❌ KMS Key not found or not accessible"
    echo "Current KMS_KEY_ID: $KMS_KEY_ID"
    echo "Listing available KMS keys..."
    aws kms list-keys --region $AWS_REGION
    echo "Please ensure you've:"
    echo "1. Run terraform successfully"
    echo "2. Updated .env with the correct KMS_KEY_ID from terraform output"
    echo "3. Have proper permissions to access the KMS key"
    exit 1
fi

echo "✅ All verifications passed!"