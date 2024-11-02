terraform {
  required_providers {
    dependencytrack = {
      source = "registry.terraform.io/solarfactories/dependencytrack"
    }
  }
}

provider "dependencytrack" {
  host = "http://localhost:8081"
  key  = "odt_dcqVqQWFy84PAxWfpEQBTItkEAMWeeoG"
}
