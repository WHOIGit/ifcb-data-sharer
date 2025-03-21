
###################
# S3 bucket with notification
###################

module "s3_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "4.1.0"

  bucket        = var.bucket_name
  force_destroy = true

  object_ownership = "BucketOwnerEnforced"

}



# create the directories for each user
resource "aws_s3_object" "folders" {
  for_each = toset(var.user_names)
  bucket   = module.s3_bucket.s3_bucket_id
  key      = "${each.value}/"
}

module "s3_notification" {
  source  = "terraform-aws-modules/s3-bucket/aws//modules/notification"
  version = "4.1.0"

  bucket = module.s3_bucket.s3_bucket_id

  eventbridge = true

  lambda_notifications = {
    lambda1 = {
      function_arn  = module.lambda_function.lambda_function_arn
      function_name = module.lambda_function.lambda_function_name
      events        = ["s3:ObjectCreated:Put", "s3:ObjectCreated:Post"]
      //filter_prefix = "data/"
      //filter_suffix = ".h5"
    }
  }
}
