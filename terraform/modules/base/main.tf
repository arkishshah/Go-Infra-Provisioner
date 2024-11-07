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

# Service policy
resource "aws_iam_role_policy" "service_policy" {
  name = "${var.project_name}-service-policy"
  role = aws_iam_role.service_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:CreateBucket",
          "s3:DeleteBucket",
          "s3:PutBucketPolicy",
          "s3:DeleteBucketPolicy",
          "s3:PutBucketVersioning",
          "s3:GetBucketLocation",
          "s3:ListBucket",
          "iam:CreateRole",
          "iam:DeleteRole",
          "iam:PutRolePolicy",
          "iam:DeleteRolePolicy",
          "iam:GetRole",
          "iam:GetRolePolicy"
        ]
        Resource = "*"
      }
    ]
  })
}
