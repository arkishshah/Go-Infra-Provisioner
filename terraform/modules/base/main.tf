provider "aws" {
  region = var.aws_region
}

# KMS key for encryption
resource "aws_kms_key" "main" {
  description             = "${var.project_name} encryption key"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  tags = {
    Environment = var.environment
    Project     = var.project_name
  }
}

# IAM role for the service
resource "aws_iam_role" "service_role" {
  name = "${var.project_name}-service-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = ["ec2.amazonaws.com", "s3.amazonaws.com"]
        }
      }
    ]
  })

  tags = {
    Environment = var.environment
    Project     = var.project_name
  }
}

# Service policy for the infrastructure provisioner
resource "aws_iam_role_policy" "service_policy" {
  name = "${var.project_name}-service-policy"
  role = aws_iam_role.service_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          # S3 permissions
          "s3:CreateBucket",
          "s3:DeleteBucket",
          "s3:PutBucketPolicy",
          "s3:GetBucketPolicy",
          "s3:PutBucketVersioning",
          "s3:GetBucketVersioning",
          "s3:GetBucketLocation",
          "s3:ListBucket",
          "s3:PutEncryptionConfiguration",
          "s3:GetEncryptionConfiguration"
        ]
        Resource = [
          "arn:aws:s3:::${var.environment}-*",
          "arn:aws:s3:::${var.environment}-*/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          # IAM permissions
          "iam:CreateRole",
          "iam:DeleteRole",
          "iam:GetRole",
          "iam:PutRolePolicy",
          "iam:DeleteRolePolicy",
          "iam:GetRolePolicy",
          "iam:ListRolePolicies",
          "iam:TagRole"
        ]
        Resource = [
          "arn:aws:iam::${var.aws_account_id}:role/${var.environment}-*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          # CloudWatch Logs permissions
          "logs:CreateLogGroup",
          "logs:DeleteLogGroup",
          "logs:PutRetentionPolicy",
          "logs:DeleteRetentionPolicy",
          "logs:DescribeLogGroups",
          "logs:TagLogGroup"
        ]
        Resource = [
          "arn:aws:logs:${var.aws_region}:${var.aws_account_id}:log-group:/aws/client/${var.environment}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          # Lambda permissions
          "lambda:CreateFunction",
          "lambda:DeleteFunction",
          "lambda:GetFunction",
          "lambda:UpdateFunctionCode",
          "lambda:UpdateFunctionConfiguration",
          "lambda:AddPermission",
          "lambda:RemovePermission",
          "lambda:TagResource"
        ]
        Resource = [
          "arn:aws:lambda:${var.aws_region}:${var.aws_account_id}:function:${var.environment}-*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          # EventBridge permissions
          "events:PutRule",
          "events:DeleteRule",
          "events:PutTargets",
          "events:RemoveTargets",
          "events:DescribeRule",
          "events:TagResource"
        ]
        Resource = [
          "arn:aws:events:${var.aws_region}:${var.aws_account_id}:rule/${var.environment}-*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          # SNS permissions
          "sns:CreateTopic",
          "sns:DeleteTopic",
          "sns:GetTopicAttributes",
          "sns:SetTopicAttributes",
          "sns:TagResource",
          "sns:Subscribe",
          "sns:Unsubscribe"
        ]
        Resource = [
          "arn:aws:sns:${var.aws_region}:${var.aws_account_id}:${var.environment}-*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          # CloudWatch Metrics & Alarms permissions
          "cloudwatch:PutMetricAlarm",
          "cloudwatch:DeleteAlarms",
          "cloudwatch:DescribeAlarms",
          "cloudwatch:TagResource"
        ]
        Resource = [
          "arn:aws:cloudwatch:${var.aws_region}:${var.aws_account_id}:alarm:${var.environment}-*"
        ]
      },
      {
        # List/Describe permissions that don't support resource-level restrictions
        Effect = "Allow"
        Action = [
          "s3:ListAllMyBuckets",
          "iam:ListRoles",
          "lambda:ListFunctions",
          "sns:ListTopics",
          "events:ListRules",
          "logs:DescribeLogGroups",
          "cloudwatch:DescribeAlarms"
        ]
        Resource = "*"
      }
    ]
  })
}
