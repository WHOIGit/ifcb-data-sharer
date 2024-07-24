# Manage external S3 users with different IAM user/policies

resource "aws_iam_user" "s3_users" {
  for_each = toset(var.user_names)
  name     = each.value
  tags = {
    Project = "${var.project_name}"
  }
}

# chain outputs from the iam_user loop
resource "aws_iam_user_policy" "s3_users" {
  for_each = aws_iam_user.s3_users
  name     = "${each.value.name}-policy"
  user     = each.value.name
  policy   = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Effect": "Allow",
      "Resource": "${module.s3_bucket.s3_bucket_arn}/${each.value.name}/*"
    }
  ]

}
EOT

}


resource "aws_iam_user" "yardon_admin" {
  name = "yarkon-admin"
  tags = {
    Project = "${var.project_name}"
  }
}

resource "aws_iam_user_policy" "yardon_admin" {
  name   = "yarkon-admin-policy"
  user   = aws_iam_user.yardon_admin.name
  policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Sid": "AllowServerToIterateBuckets",
        "Effect": "Allow",
        "Action": "s3:ListAllMyBuckets",
        "Resource": "arn:aws:s3:::*"
    }, 
    {
        "Sid": "AllowServerToAccessSpecificBuckets",
        "Effect": "Allow",
        "Action": [
            "s3:ListBucket",
            "s3:GetBucketLocation",
            "s3:GetBucketCORS",
            "s3:PutBucketCORS"
        ],
        "Resource": [
            "${module.s3_bucket.s3_bucket_arn}"
        ]
    }, {
        "Sid": "AllowUserActionsLimitedToSpecificBuckets",
        "Effect": "Allow",
        "Action": "s3:*",
        "Resource": [
            "${module.s3_bucket.s3_bucket_arn}/*"
        ]
    }
  ]

}
EOT

}
