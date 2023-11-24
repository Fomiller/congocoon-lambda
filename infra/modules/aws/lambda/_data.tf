data "archive_file" "zip" {
  type        = "zip"
  source_file = "${path.module}/src/bin/lambda_congocoon"
  output_path = "${path.module}/lambda_function.zip"
}

data "aws_kms_key" "fomiller_master" {
  key_id = "alias/${var.namespace}-master"
}

data "aws_secretsmanager_secret" "gmail_api_key" {
  name = "${var.namespace}-gmail-api-key"
}

