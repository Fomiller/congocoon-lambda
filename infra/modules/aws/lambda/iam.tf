resource "aws_iam_role" "lambda_role" {
  name               = "${title(var.namespace)}LambdaCongocoon"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "lambda_role_policy" {
  name        = "${title(var.namespace)}LambdaCongocoonPolicy"
  description = "IAM Policy for ${title(var.namespace)}LambdaCongocoon"
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [{
      "Sid" : "AllowCloudwatchLogging",
      "Effect" : "Allow",
      "Action" : [
        "logs:CreateLogStream",
        "logs:CreateLogDelivery",
        "logs:PutLogEvents"
      ],
      "Resource" : "arn:aws:logs:*:*:*"
      },
      {
        "Sid" : "ListObjectsInBucket",
        "Effect" : "Allow",
        "Action" : ["s3:ListBucket"],
        "Resource" : ["arn:aws:s3:::${var.namespace}-dev"]
      },
      {
        "Sid" : "AllObjectActions",
        "Effect" : "Allow",
        "Action" : "s3:*Object",
        "Resource" : ["arn:aws:s3:::${var.namespace}-dev/*"]
      },
      {
        "Sid" : "GetGmailSecret",
        "Effect" : "Allow",
        "Action" : "secretsmanager:GetSecretValue",
        "Resource" : [data.aws_secretsmanager_secret.gmail_api_key.arn]
      },
      {
        "Sid" : "ListDescribeSecret",
        "Effect" : "Allow",
        "Action" : "secretsmanager:DescribeSecret",
        "Action" : "secretsmanager:List*",
        "Resource" : ["*"]
      },
      {
        "Sid" : "KmsDecrypt",
        "Effect" : "Allow",
        "Action" : "kms:Decrypt",
        "Resource" : [data.aws_kms_key.fomiller_master.arn]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_role_policy.arn
}

