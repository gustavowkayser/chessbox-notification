output "api_endpoint" {
  description = "The public invoke URL of the API Gateway HTTP API"
  value       = aws_apigatewayv2_stage.default.invoke_url
}

output "api_lambda_name" {
  description = "Name of the API Lambda function"
  value       = aws_lambda_function.api.function_name
}

output "worker_lambda_name" {
  description = "Name of the Worker Lambda function"
  value       = aws_lambda_function.worker.function_name
}

output "app_secret" {
  description = "Application secret / Argon2 salt"
  value       = local.app_secret
  sensitive   = true
}
