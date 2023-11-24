resource "aws_lambda_function" "lambda" {
  function_name    = "${var.namespace}-${var.app_prefix}-scraper"
  role             = aws_iam_role.lambda_role.arn
  filename         = "${path.module}/lambda_function.zip"
  handler          = "bootstrap"
  source_code_hash = data.archive_file.zip.output_base64sha256
  runtime          = "provided.al2"
  architectures    = ["arm64"]
  memory_size      = 128
  timeout          = 10
}

resource "aws_lambda_permission" "lambda_permission" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.lambda.arn
}

