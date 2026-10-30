# EventBridge Rule for periodic tournament worker execution
resource "aws_cloudwatch_event_rule" "worker_schedule" {
  name                = "${local.name_prefix}-worker-schedule"
  description         = "Periodic trigger for tournament scraper and push notification worker"
  schedule_expression = var.worker_schedule

  tags = {
    Name = "${local.name_prefix}-worker-schedule"
  }
}

# Target: Worker Lambda Function
resource "aws_cloudwatch_event_target" "worker_lambda_target" {
  rule      = aws_cloudwatch_event_rule.worker_schedule.name
  target_id = "WorkerLambda"
  arn       = aws_lambda_function.worker.arn
}

# Permission for EventBridge to invoke the Worker Lambda
resource "aws_lambda_permission" "eventbridge" {
  statement_id  = "AllowEventBridgeInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.worker.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.worker_schedule.arn
}
