resource "dependencytrack_team" "example" {
  name = "Example"
}

resource "dependencytrack_team_permission" "example" {
  team       = dependencytrack_team.id
  permission = "BOM_UPLOAD"
}
