resource "dependencytrack_policy" "example" {
  name      = "Sample Policy"
  operator  = "ALL"
  violation = "ERROR"
}

resource "dependencytrack_policy_tag" "example" {
  policy = dependencytrack_policy.example.id
  tag    = "DemoTag"
}
