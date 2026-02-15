data "dependencytrack_oidc_login" "example" {
  id_token = "eyJ..."
}

output "oidc_available" {
  value = data.dependencytrack_oidc_login.example.token
}
