resource "dependencytrack_team" "example" {
  name = "Example Team Name"
}

resource "dependencytrack_oidc_group" "example" {
  name = "Example Group Name"
}

resource "dependencytrack_oidc_group_mapping" "example" {
  group = dependencytrack_oidc_group.example.id
  team  = dependencytrack_oidc_group.example.team
}
