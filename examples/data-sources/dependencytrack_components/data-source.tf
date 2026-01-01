data "dependencytrack_project" "example" {
  name    = "Example"
  version = "v1"
}

// All Components
data "dependencytrack_components" "example" {
  project = data.dependencytrack_project.example.id
}

// Filtered
data "dependencytrack_components" "filtered" {
  project       = data.dependencytrack_project.example.id
  only_direct   = true
  only_outdated = true
}
