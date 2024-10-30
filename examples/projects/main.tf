terraform {
	required_providers {
		dependencytrack = {
			source = "registry.terraform.io/solarfactories/dependencytrack"
		}
	}
}

provider "dependencytrack" {
	host = "http://localhost:8081"
	token = "odt_dcqVqQWFy84PAxWfpEQBTItkEAMWeeoG"
}

// Requires the creation of a project within DependencyTrack.
// By default, DependencyTrack defaults to an empty Version,
// which prevents the Go SDK from being able to find it with `Lookup`
data "dependencytrack_project" "example" {
	name = "Example"
	version = "v1"
}
