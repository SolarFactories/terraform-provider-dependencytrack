data "dependencytrack_user_login" "example" {
  username = "admin"
  password = "admin"
}

output "bearer_token" {
  value = data.dependencytrack_user_login.example.token
}
