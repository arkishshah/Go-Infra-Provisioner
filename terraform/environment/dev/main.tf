terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

module "base" {
  source = "../../modules/base"

  aws_region   = var.aws_region
  environment  = var.environment
  project_name = var.project_name
  aws_account_id = var.aws_account_id
}

output "service_role_arn" {
  value = module.base.service_role_arn
}

output "kms_key_id" {
  value = module.base.kms_key_id
}
