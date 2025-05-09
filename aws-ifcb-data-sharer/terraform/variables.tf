variable "project_name" {
  description = "Project name"
  default     = "ifcb-data-sharer"
}

variable "bucket_name" {
  description = "S3 bucket name"
  default     = "ifcb-data-sharer.files"
}

variable "user_names" {
  description = "Create IAM users with these names"
  type        = list(string)
  default     = ["eandrews", "hablab", "aoos", "nwfsc", "bowdoin", "umaine", "unh", "vims", "fwri", "stonybrook", "tamu", "osu", "smhi", "laval"]
}

variable "aws_account_id" {
  description = "AWS Account ID"
  sensitive   = true
}
variable "ecr_root" {
  description = "ECR root url"
  default     = "139464377685.dkr.ecr.us-east-1.amazonaws.com"
}
