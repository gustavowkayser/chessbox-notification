# Fallback dummy zip if build/bootstrap.zip does not yet exist
data "archive_file" "dummy" {
  type        = "zip"
  output_path = "${path.module}/dummy_bootstrap.zip"

  source {
    content  = "#!/bin/sh\necho 'placeholder bootstrap'"
    filename = "bootstrap"
  }
}

# Generate application secret (salt for Argon2) if not provided
resource "random_password" "app_secret" {
  length  = 32
  special = false
}

locals {
  zip_exists     = fileexists("${path.module}/${var.lambda_zip_path}")
  lambda_payload = local.zip_exists ? "${path.module}/${var.lambda_zip_path}" : data.archive_file.dummy.output_path
  lambda_hash    = local.zip_exists ? filebase64sha256("${path.module}/${var.lambda_zip_path}") : data.archive_file.dummy.output_base64sha256
  app_secret     = var.app_secret != "" ? var.app_secret : random_password.app_secret.result
}

# CloudWatch Log Group for API Lambda
resource "aws_cloudwatch_log_group" "api" {
  name              = "/aws/lambda/${local.name_prefix}-api"
  retention_in_days = 14
}

# API Lambda Function (Handles API Gateway HTTP requests)
resource "aws_lambda_function" "api" {
  function_name    = "${local.name_prefix}-api"
  role             = aws_iam_role.lambda_exec.arn
  runtime          = "provided.al2023"
  architectures    = [var.lambda_architecture]
  handler          = "bootstrap"
  filename         = local.lambda_payload
  source_code_hash = local.lambda_hash
  timeout          = 30
  memory_size      = 256

  environment {
    variables = {
      DATABASE_URL = var.database_url
      SECRET       = local.app_secret
      APP_MODE     = "api"
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic,
    aws_cloudwatch_log_group.api
  ]

  tags = {
    Name = "${local.name_prefix}-api"
    Role = "api"
  }
}

# CloudWatch Log Group for Worker Lambda
resource "aws_cloudwatch_log_group" "worker" {
  name              = "/aws/lambda/${local.name_prefix}-worker"
  retention_in_days = 14
}

# Worker Lambda Function (Periodic scraper and push notification dispatcher)
resource "aws_lambda_function" "worker" {
  function_name    = "${local.name_prefix}-worker"
  role             = aws_iam_role.lambda_exec.arn
  runtime          = "provided.al2023"
  architectures    = [var.lambda_architecture]
  handler          = "bootstrap"
  filename         = local.lambda_payload
  source_code_hash = local.lambda_hash
  timeout          = 60
  memory_size      = 256

  environment {
    variables = {
      DATABASE_URL = var.database_url
      SECRET       = local.app_secret
      APP_MODE     = "worker"
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic,
    aws_cloudwatch_log_group.worker
  ]

  tags = {
    Name = "${local.name_prefix}-worker"
    Role = "worker"
  }
}
