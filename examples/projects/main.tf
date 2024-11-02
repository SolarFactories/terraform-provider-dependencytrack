terraform {
  required_providers {
    dependencytrack = {
      source = "registry.terraform.io/solarfactories/dependencytrack"
    }
  }
}

provider "dependencytrack" {
  host = "http://localhost:8081"
  key  = "OS_ENV"
}

// Requires the creation of a project within DependencyTrack.
// By default, DependencyTrack defaults to an empty Version,
// which prevents the Go SDK from being able to find it with `Lookup`
data "dependencytrack_project" "example" {
  name    = "Example"
  version = "v1"
}

resource "dependencytrack_project" "example" {
  name = "Example"
}

resource "dependencytrack_project" "example2" {
  name        = "Example 2"
  description = "A Sample project as generated by terraform"
  active      = true
}

output "project_example_data" {
  value = data.dependencytrack_project.example
}

output "project_example_resource" {
  value = dependencytrack_project.example2
}