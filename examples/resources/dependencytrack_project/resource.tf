resource "dependencytrack_project" "example" {
  name        = "Example"
  description = "Example project"
}

// Project collecting - API v4.13+
resource "dependencytrack_project" "example_collection" {
  name        = "Example Collection"
  description = "Example Collection Project"
  collection = {
    logic = "AGGREGATE_DIRECT_CHILDREN"
  }
}
