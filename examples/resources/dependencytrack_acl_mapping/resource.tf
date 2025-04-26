resource "dependencytrack_project" "example" {
  name = "Example Project"
}

resource "dependencytrack_team" "example" {
  name = "Example Team"
}

resource "dependencytrack_acl_mapping" "example" {
  team    = dependencytrack_team.example.id
  project = dependencytrack_project.example.id
}
