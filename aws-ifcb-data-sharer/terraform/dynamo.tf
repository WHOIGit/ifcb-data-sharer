module "dynamodb_table" {
  source = "terraform-aws-modules/dynamodb-table/aws"

  name                        = "ifcb-data-sharing"
  hash_key                    = "user"
  range_key                   = "pid"
  table_class                 = "STANDARD"
  deletion_protection_enabled = false
  stream_enabled              = true
  stream_view_type            = "NEW_IMAGE"
  attributes = [
    {
      name = "user"
      type = "S"
    },
    {
      name = "pid"
      type = "S"
    },
  ]

  tags = {
    Terraform   = "true"
    Environment = "staging"
  }
}
