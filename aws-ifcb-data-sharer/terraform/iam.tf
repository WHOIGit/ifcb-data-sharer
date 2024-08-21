# IAM User for full S3 access
resource "aws_iam_user" "prod_bucket" {
  name = "${var.project_name}-s3-bucket"
}

resource "aws_iam_user_policy" "prod_bucket" {
  user = aws_iam_user.prod_bucket.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:*",
        ]
        Effect = "Allow"
        Resource = [
          "${module.s3_bucket.s3_bucket_arn}",
          "${module.s3_bucket.s3_bucket_arn}/*"
        ]
      },
    ]
  })
}

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
      "Effect": "Allow",
      "Action": [
        "s3:ListBucket"
      ],
      "Resource": "${module.s3_bucket.s3_bucket_arn}"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject"
      ],
      "Resource": "${module.s3_bucket.s3_bucket_arn}/${each.value.name}/*"
    },
    {
      "Effect": "Allow",
      "Action": "s3:PutObject",
      "Resource": "${module.s3_bucket.s3_bucket_arn}/${each.value.name}/*"
      
    }
  ]

}
EOT
}

# create the group and policies for each user to use Yarkon
resource "aws_iam_group" "s3_users_yarkon" {
  for_each = aws_iam_user.s3_users
  name     = "${each.value.name}-yarkon"
}

resource "aws_iam_group_policy" "s3_users_yarkon" {
  for_each = aws_iam_user.s3_users
  name     = "${each.value.name}-yarkon-policy"
  group    = "${each.value.name}-yarkon"

  policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Sid": "AllowServerToAccessSpecificBuckets",
        "Effect": "Allow",
        "Action": [
            "s3:ListBucket",
            "s3:GetBucketLocation"
        ],
        "Resource": "${module.s3_bucket.s3_bucket_arn}"
    }, 
    {
        "Sid": "AllowUserActionsLimitedToSpecificBuckets",
        "Effect": "Allow",
        "Action": "s3:*",
        "Resource": "${module.s3_bucket.s3_bucket_arn}/${each.value.name}/*"
    }
  ]

}
EOT
}

resource "aws_iam_group" "my_developers" {
  name = "developers"
  path = "/users/"
}

# Admin Yarkon resrouce
resource "aws_iam_user" "yardon_admin" {
  name = "yarkon-admin"
  tags = {
    Project = "${var.project_name}"
  }
}
resource "aws_iam_user_policy_attachment" "yarkon_admin" {
  user       = aws_iam_user.yardon_admin.name
  policy_arn = aws_iam_policy.yarkon_admin.arn
}

resource "aws_iam_role" "yarkon_admin" {
  name               = "yarkon-admin-role"
  assume_role_policy = <<EOT
  {
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "ec2.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        },
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": "arn:aws:iam::${var.aws_account_id}:root"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
EOT

  inline_policy {
    name = "yarkon-inline-policy"

    policy = <<EOT
    {
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowServerToIterateIAMEntities",
            "Effect": "Allow",
            "Action": [
                "iam:Get*",
                "iam:List*"
            ],
            "Resource": "arn:aws:iam::${var.aws_account_id}:*"
        }, 
        {
            "Sid": "AllowServerToAssumeRole",
            "Effect": "Allow",
            "Action": "sts:AssumeRole",
            "Resource": "arn:aws:iam::${var.aws_account_id}:role/yarkon-admin-role"
        },
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
}
resource "aws_iam_policy" "yarkon_admin" {
  name = "yarkon-admin-policy"

  policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [
    {
        "Sid": "AllowServerToIterateIAMEntities",
        "Effect": "Allow",
        "Action": [
            "iam:Get*",
            "iam:List*"
        ],
        "Resource": "arn:aws:iam::${var.aws_account_id}:*"
    }, 
    {
        "Sid": "AllowServerToAssumeRole",
        "Effect": "Allow",
        "Action": "sts:AssumeRole",
        "Resource": "arn:aws:iam::${var.aws_account_id}:role/yarkon-admin-role"
    },
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
