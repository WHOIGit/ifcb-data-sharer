
###################
# S3 bucket with notification
###################

module "s3_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "4.1.0"

  bucket        = var.bucket_name
  force_destroy = true

  object_ownership = "BucketOwnerEnforced"

  #attach_policy = true
  # bucket policy to limit uploads to only IFCB file extensions

}

# create the directories for each user
resource "aws_s3_object" "folders" {
  for_each = toset(var.user_names)
  bucket   = module.s3_bucket.s3_bucket_id
  key      = "${each.value}/"
}
