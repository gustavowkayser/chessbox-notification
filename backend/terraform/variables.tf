variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (e.g. production, staging, dev)"
  type        = string
  default     = "production"
}

variable "app_name" {
  description = "Application name"
  type        = string
  default     = "chessbox-notification"
}

variable "lambda_architecture" {
  description = "Lambda CPU architecture (arm64 or x86_64)"
  type        = string
  default     = "arm64"
  validation {
    condition     = contains(["arm64", "x86_64"], var.lambda_architecture)
    error_message = "lambda_architecture must be either 'arm64' or 'x86_64'."
  }
}

variable "lambda_zip_path" {
  description = "Relative path to the pre-packaged Lambda zip file"
  type        = string
  default     = "../build/bootstrap.zip"
}

variable "database_url" {
  description = "Connection URL for external PostgreSQL database provider (e.g., Neon, Supabase, CockroachDB)"
  type        = string
  sensitive   = true
}

variable "app_secret" {
  description = "Argon2 salt/secret key for device credentials hashing (leave empty to generate securely via Terraform)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "worker_schedule" {
  description = "EventBridge cron schedule for the tournament scraper worker"
  type        = string
  default     = "rate(1 minute)"
}
