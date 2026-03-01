data "dependencytrack_oidc_available" "example" {}

output "oidc_available" {
  value = data.dependencytrack_oidc_available.available
}
