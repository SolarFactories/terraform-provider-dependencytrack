data "dependencytrack_oidc_users" "example" {}

output "count" {
  value = data.dependencytrack_oidc_users.example.total_count
}

output "users" {
  value = data.dependencytrack_oidc_users.example.users
}
