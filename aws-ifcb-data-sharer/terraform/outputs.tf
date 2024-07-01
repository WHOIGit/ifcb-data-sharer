output "all_users" {
  value = aws_iam_user.s3_users
}

output "all_policies" {
  value = aws_iam_user_policy.s3_users
}
