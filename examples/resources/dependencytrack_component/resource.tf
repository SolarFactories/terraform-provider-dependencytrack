resource "dependencytrack_project" "example" {
  name        = "Example"
  description = "Example project"
}

resource "dependencytrack_component" "example" {
  project = dependencytrack_project.example.id
  name    = "ComponentName"
  version = "v1.0.0"
  hashes  = {}
}
