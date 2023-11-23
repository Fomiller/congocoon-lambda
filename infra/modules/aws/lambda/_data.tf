data "archive_file" "zip" {
  type        = "zip"
  source_file = "${path.module}/src/bin/lambda_congocoon"
  output_path = "${path.module}/lambda_function.zip"
}

data "aws_kms_key" "chat_stat_master_kms_key" {
  key_id = "alias/${var.namespace}-chat-stat-master"
}

data "aws_secretsmanager_secret" "gmail_api_key" {
  name = "${var.namespace}-gmail-api-key"
}

