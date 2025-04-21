resource "dependencytrack_team" "example" {
  name = "Example Team Name"
}

resource "dependencytrack_oidc_group" "example" {
  name = "Example Group Name"
}

resource "dependencytrack_config_property" "example" {
  group = dependencytrack_oidc_group.example.id
  team  = dependencytrack_oidc_group.example.team
}
