resource "dependencytrack_team" "example" {
  name = "Example"
}

resource "dependencytrack_oidc_group" "example" {
  name = "Example"
}

resource "dependencytrack_oidc_group_mapping" "example" {
  group = dependencytrack_oidc_group.example.id
  team  = dependencytrack_team.example.id
}

data "dependencytrack_oidc_group_mappings" "example" {
  group      = dependencytrack_oidc_group.example.id
  depends_on = dependencytrack_oidc_group_mapping.example
}
