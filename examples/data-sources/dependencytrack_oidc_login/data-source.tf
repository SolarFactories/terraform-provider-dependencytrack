data "dependencytrack_oidc_login" "example" {
  id_token = "eyJ..."
}

output "bearer_token" {
  value = data.dependencytrack_oidc_login.example.token
}
