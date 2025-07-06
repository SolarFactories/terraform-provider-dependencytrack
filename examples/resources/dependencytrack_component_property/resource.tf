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

resource "dependencytrack_component_property" "example" {
  component   = dependencytrack_component.example.id
  group       = "PropertyGroup"
  name        = "PropertyName"
  value       = "PropertyValue"
  type        = "STRING"
  description = "PropertyDescription"
}
