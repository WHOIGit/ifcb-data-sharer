

module "dynamodb_table_sharer" {
  source = "terraform-aws-modules/dynamodb-table/aws"

  name                        = "ifcb-data-sharer-bins"
  hash_key                    = "username"
  range_key                   = "pid"
  table_class                 = "STANDARD"
  deletion_protection_enabled = false
  stream_enabled              = true
  stream_view_type            = "NEW_IMAGE"
  attributes = [
    {
      name = "username"
      type = "S"
    },
    {
      name = "pid"
      type = "S"
    },
  ]

  tags = {
    Terraform = "true"
  }
}
