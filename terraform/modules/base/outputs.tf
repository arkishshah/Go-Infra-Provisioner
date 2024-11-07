output "service_role_arn" {
  description = "ARN of the service IAM role"
  value       = aws_iam_role.service_role.arn
}

output "kms_key_id" {
  description = "ID of the KMS key"
  value       = aws_kms_key.main.id
}
