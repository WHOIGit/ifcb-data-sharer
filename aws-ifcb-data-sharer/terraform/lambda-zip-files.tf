module "lambda_function_zip" {
  source = "terraform-aws-modules/lambda/aws"

  function_name = "zip-ifcb-files"
  description   = "Zip three IFCB files together if DynamoDB says they're validated"
  handler       = "app.lambda_handler"
  runtime       = "python3.12"
  publish       = true

  source_path = "${path.module}/../lambdas/zip-ifcb-files"

  # architecture config
  memory_size = 256
  timeout     = 300

  event_source_mapping = {
    dynamodb = {
      event_source_arn  = module.dynamodb_table.dynamodb_table_stream_arn
      starting_position = "LATEST"

      filter_criteria = [
        {
          pattern = jsonencode({
            eventName : ["INSERT", "MODIFY"]
          })
        },
      ]
    }
  }
  # cloudwatch
  cloudwatch_logs_retention_in_days = 7

  # role and policy config
  attach_policy_statements = true
  policy_statements = {
    DynamoStrem = {
      effect  = "Allow",
      actions = ["dynamodb:GetItem", "dynamodb:PutItem", "dynamodb:UpdateItem", "dynamodb:Query"],
      resources = [
        module.dynamodb_table.dynamodb_table_arn
      ]
    }
  }

  attach_policies    = true
  number_of_policies = 1

  policies = [
    "arn:aws:iam::aws:policy/service-role/AWSLambdaDynamoDBExecutionRole",
  ]
}
