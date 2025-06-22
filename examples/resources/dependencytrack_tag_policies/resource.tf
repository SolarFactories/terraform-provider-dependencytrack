# Requires DependencyTrack API v4.12+

resource "dependencytrack_project" "example" {
  name = "Example Project"
  tags = ["example_tag"]
}

resource "dependencytrack_policy" "example" {
  name      = "Example Policy"
  operator  = "ANY"
  violation = "FAIL"
}

resource "dependencytrack_tag_policies" "example" {
  tag = "example_tag"
  policies = [
    dependencytrack_policy.example.id,
  ]
  depends_on = [dependencytrack_project.example]
}
