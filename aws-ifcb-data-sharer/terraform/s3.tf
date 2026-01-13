
###################
# S3 bucket with notification
###################
data "aws_caller_identity" "current" {}

module "s3_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "4.1.0"

  bucket        = var.bucket_name
  force_destroy = true

  object_ownership = "BucketOwnerEnforced"
  logging = {
    target_bucket = module.log_bucket.s3_bucket_id
    target_prefix = "log/"
    target_object_key_format = {
      partitioned_prefix = {
        partition_date_source = "DeliveryTime" # "EventTime"
      }
      # simple_prefix = {}
    }
  }

}

module "log_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "4.1.0"

  bucket        = "logs-ifcb-data-sharer"
  force_destroy = true

  control_object_ownership = true

  attach_access_log_delivery_policy = true

  access_log_delivery_policy_source_accounts = [data.aws_caller_identity.current.account_id]
  access_log_delivery_policy_source_buckets  = ["arn:aws:s3:::${var.bucket_name}"]
  lifecycle_rule = [
    {
      id     = "log-archive-rule"
      status = "Enabled"

      # Rule applies to objects with the "logs/" prefix

      filter = {
        prefix = "log/"
      }


      # Transition objects to STANDARD_IA (Infrequent Access) after 30 days
      /*
      transition = {
        days          = 7
        storage_class = "STANDARD_IA"
      }
      */

      # Permanently delete objects after 90 days
      expiration = {
        days = 14
      }
    }
  ]
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
      events        = ["s3:ObjectCreated:Put", "s3:ObjectCreated:Post", "s3:ObjectCreated:CompleteMultipartUpload"]
      //filter_prefix = "data/"
      //filter_suffix = ".h5"
    }
  }
}
