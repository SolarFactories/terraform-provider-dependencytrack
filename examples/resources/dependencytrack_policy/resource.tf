resource "dependencytrack_policy" "example" {
  name      = "Sample Policy"
  operator  = "ALL"
  violation = "ERROR"
}
