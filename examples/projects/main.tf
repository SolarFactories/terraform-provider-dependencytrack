terraform {
	required_providers {
		dependencytrack = {
			source = "registry.terraform.io/solarfactories/dependencytrack"
		}
	}
}

provider "dependencytrack" {
	host = "localhost:8080"
	token = "odt_dcqVqQWFy84PAxWfpEQBTItkEAMWeeoG"
}

data "dependencytrack_project" "example" {}
