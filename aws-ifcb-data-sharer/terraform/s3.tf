###################
# S3 bucket with notification
###################

module "s3_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "4.1.0"

  bucket        = "ifcb-data-sharer.files"
  force_destroy = true
  acl           = "private"

  control_object_ownership = true
  object_ownership         = "ObjectWriter"
}
