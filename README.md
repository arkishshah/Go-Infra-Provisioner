# AWS Infrastructure Provisioner

A Go-based service that automates the provisioning of AWS resources (S3 buckets and IAM roles) for client organizations. This service provides a REST API to dynamically create and manage AWS infrastructure with proper access controls and permissions.

## Features

- 🚀 Automated AWS resource provisioning
- 🔐 Secure authentication and authorization
- 🏗️ Infrastructure as Code using Terraform
- 📝 REST API endpoints for resource management
- 🧪 Testing and verification scripts
- 🧹 Resource cleanup utilities

## Prerequisites

Before you begin, ensure you have the following installed:
- [Go](https://golang.org/doc/install) (version 1.21 or later)
- [Terraform](https://www.terraform.io/downloads.html) (version 1.0.0 or later)
- [AWS CLI](https://aws.amazon.com/cli/) configured with appropriate credentials
- [Git](https://git-scm.com/downloads)

## AWS Setup

1. Create an AWS Account if you don't have one
2. Create an IAM user with programmatic access:
   - Go to AWS Console → IAM → Users → Add User
   - Enable programmatic access
   - Attach the following permissions policy:

```json
 {
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "TerraformS3Access",
            "Effect": "Allow",
            "Action": [
                "s3:CreateBucket",
                "s3:ListBucket",
                "s3:GetBucketPolicy",
                "s3:PutBucketPolicy",
                "s3:DeleteBucket"
            ],
            "Resource": "arn:aws:s3:::*"
        },
        {
            "Sid": "TerraformIAMAccess",
            "Effect": "Allow",
            "Action": [
                "iam:CreateRole",
                "iam:GetRole",
                "iam:DeleteRole",
                "iam:PutRolePolicy",
                "iam:GetRolePolicy",
                "iam:DeleteRolePolicy",
                "iam:ListRoles",
                "iam:ListRolePolicies",
                "iam:TagRole",
                "iam:ListAttachedRolePolicies",
                "iam:ListInstanceProfilesForRole"
            ],
            "Resource": [
                "arn:aws:iam::*:role/go-infra-provisioner-*"
            ]
        },
        {
            "Sid": "TerraformKMSAccess",
            "Effect": "Allow",
            "Action": [
                "kms:CreateKey",
                "kms:DescribeKey",
                "kms:EnableKeyRotation",
                "kms:ListKeys",
                "kms:PutKeyPolicy",
                "kms:GenerateDataKey",
                "kms:TagResource",
                "kms:GetKeyRotationStatus",
                "kms:GetKeyPolicy",
                "kms:ListResourceTags",
                "kms:ScheduleKeyDeletion"
            ],
            "Resource": "*"
        }
    ]
}
   ```
3. Save the Access Key ID and Secret Access Key

## Installation

1. Clone the repository:
```bash
git clone https://github.com/ashah/go-infra-provisioner.git
cd go-infra-provisioner
```

2. Create environment files:

Create `.env` file in the root directory:
```plaintext
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_REGION=us-east-1
AWS_ACCOUNT_ID=your_account_id
ENVIRONMENT=dev
```

Create `configs/dev/app.env` with the same contents:
```bash
mkdir -p configs/dev
cp .env configs/dev/app.env
```

3. Install dependencies:
```bash
go mod download
```

## Infrastructure Setup

1. Navigate to the terraform environment directory:
```bash
cd terraform/environments/dev
```

2. Create terraform.tfvars file:
```bash
# Copy example.tfvars to terraform.tfvars
cp example.tfvars terraform.tfvars
```

3. Update terraform.tfvars with your values:
```hcl
aws_region = "us-east-1"      # Your AWS region
environment = "dev"           # Environment name
project_name = "go-infra-provisioner"  # Your project name
```

4. Initialize Terraform:
```bash
terraform init
```

5. Review and apply the configuration:
```bash
terraform plan
terraform apply
```

6. After successful apply, you'll see outputs like:
```bash
Outputs:
service_role_arn = "arn:aws:iam::your_account_id:role/infra-provisioner-service-role"
kms_key_id = "12345678-abcd-efgh-ijkl-123456789012"
`````

7. Update your `.env` and `configs/dev/app.env` with these values.

> **Note**: The `terraform.tfvars` file contains sensitive configuration and is excluded from git via .gitignore. The `example.tfvars` file is provided as a template.

## Running the Service

1. Build and run the service:
```bash
go build -o main cmd/api/main.go
./main
```

Or use the provided Makefile:
```bash
make run
```

2. The service will start on `http://localhost:8080`

## API Endpoints

1. Health Check:
```bash
curl http://localhost:8080/health
```

2. Provision Resources:
```bash
curl -X POST http://localhost:8080/api/v1/provision \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "test-client-001",
    "client_name": "Test Client"
  }'
```

## Testing

1. Verify setup:
```bash
./scripts/verify-setup.sh
```

2. Test API endpoints:
```bash
./scripts/test-api.sh
```

Or use the Makefile:
```bash
make test
```

## Cleanup

To clean up resources:
```bash
# Clean up specific client resources
./scripts/cleanup-resources.sh test-client-001

# Clean up all terraform resources
make clean
```

## Project Structure
```
.
├── cmd/
│   └── api/                  # Application entrypoint
├── configs/
│   └── dev/                  # Environment configurations
├── internal/
│   ├── api/                  # API implementation
│   ├── config/               # Configuration management
│   └── provisioner/          # AWS resource provisioning
├── pkg/
│   ├── awsclient/           # AWS SDK client
│   └── logger/              # Logging utility
├── policies/                # IAM policy templates
├── scripts/                 # Utility scripts
├── terraform/               # Infrastructure as Code
│   ├── environments/        # Environment-specific configs
│   └── modules/            # Reusable terraform modules
├── .env                     # Environment variables
├── go.mod                   # Go dependencies
└── Makefile                # Build automation
```

## Common Issues

1. **AWS Region Error**:
   - Ensure AWS_REGION in .env matches your AWS CLI configuration
   - For non us-east-1 regions, update the S3 bucket creation configuration

2. **Permission Issues**:
   - Verify IAM user has necessary permissions
   - Check if role/policy names conflict with existing resources

3. **Resource Limits**:
   - Be aware of AWS service limits
   - Clean up test resources after use

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details

## Support

For support, please open an issue in the GitHub repository or contact the maintainers.
