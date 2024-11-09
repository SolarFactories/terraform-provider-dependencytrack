resource "dependencytrack_project" "example" {
  name        = "Example"
  description = "Example project"
}

resource "dependencytrack_project_property" "example" {
  project = dependencytrack_project.example.id
  group   = "GroupName"
  name    = "PropertyName"
}
