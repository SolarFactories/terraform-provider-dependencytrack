data "dependencytrack_oidc_users" "example" {}

output "count" {
  value = dependencytrack_oidc_users.example.total_count
}

output "users" {
  value = dependencytrack_oidc_users.example.users
}
