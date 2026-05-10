data "dependencytrack_permissions" "example" {}

output "permissions" {
  value = data.dependencytrack_permissions.example.permissions
}
