terraform {
	required_providers {
		dependencytrack = {
			source = "registry.terraform.io/solarfactories/dependencytrack"
		}
	}
}

provider "dependencytrack" {}

data "dependencytrack_project" "example" {}
