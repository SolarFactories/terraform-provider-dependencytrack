data "dependencytrack_project" "example" {
  name    = "Example"
  version = "v1"
}

data "dependencytrack_project_property" "example" {
  project = data.dependencytrack_project.example.id
  group   = "GroupName"
  name    = "PropertyName"
}
