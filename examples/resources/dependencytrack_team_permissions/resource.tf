resource "dependencytrack_team" "example" {
  name = "Example"
}

resource "dependencytrack_team_permissions" "example" {
  team        = dependencytrack_team.example.id
  permissions = ["BOM_UPLOAD"]
}
