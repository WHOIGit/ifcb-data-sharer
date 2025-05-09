data "aws_ecr_authorization_token" "token" {}

provider "docker" {
  registry_auth {
    address  = var.ecr_root
    username = data.aws_ecr_authorization_token.token.user_name
    password = data.aws_ecr_authorization_token.token.password
  }
}

module "docker_image" {
  source = "terraform-aws-modules/lambda/aws//modules/docker-build"

  create_ecr_repo = true
  ecr_repo        = "validate-ifcb-files-lambda"

  use_image_tag = true
  image_tag     = "1.21"

  source_path = "${path.module}/../lambdas/validate-ifcb-files"

}

#############################################
# Lambda Function (from image)
#############################################

module "lambda_function" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "7.7.1"

  function_name  = "validate-ifcb-files-lambda"
  description    = "Validate all files that uploaded by users, only allow valid ADC, HDR and ROI files"
  create_package = false
  publish        = true

  # architecture config
  memory_size = 256
  timeout     = 300
  # throttle lambda execution to not kill habon-ifcb api with requests
  reserved_concurrent_executions = 100

  # container config
  image_uri     = module.docker_image.image_uri
  package_type  = "Image"
  architectures = ["x86_64"]

  allowed_triggers = {
    AllowExecutionFromS3Bucket = {
      service    = "s3"
      source_arn = module.s3_bucket.s3_bucket_arn
    }
  }
  # cloudwatch
  cloudwatch_logs_retention_in_days = 7

  # role and policy config
  attach_policy_statements = true
  policy_statements = {
    GetS3Objects = {
      effect  = "Allow",
      actions = ["s3:GetObject", "s3:DeleteObject", "s3:CopyObject", "s3:PutObject", "s3:ListBucket"],
      resources = [
        "${module.s3_bucket.s3_bucket_arn}",
        "${module.s3_bucket.s3_bucket_arn}/*"
      ]
    },
    DynamoDb = {
      effect  = "Allow",
      actions = ["dynamodb:GetItem", "dynamodb:PutItem", "dynamodb:UpdateItem", "dynamodb:Query"],
      resources = [
        module.dynamodb_table_sharer.dynamodb_table_arn
      ]
    }

  }
}
