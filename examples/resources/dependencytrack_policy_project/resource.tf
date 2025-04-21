resource "dependencytrack_policy" "example" {
  name      = "Sample Policy"
  operator  = "ALL"
  violation = "ERROR"
}

resource "dependencytrack_project" "example" {
  name = "Example Project"
}

resource "dependencytrack_policy_project" "example" {
  policy  = dependencytrack_policy.example.id
  project = dependencytrack_project.example.id
}
