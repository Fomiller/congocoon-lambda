resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${aws_lambda_function.lambda.function_name}"
  retention_in_days = 7
}

resource "aws_cloudwatch_event_rule" "lambda" {
  name                = "congocoon-lambda-interval"
  description         = "Trigger lambda every hour"
  schedule_expression = "rate(30 minutes)"
}

resource "aws_cloudwatch_event_target" "lambda" {
  rule      = aws_cloudwatch_event_rule.lambda.name
  target_id = "lambda"
  arn       = aws_lambda_function.lambda.arn
}


