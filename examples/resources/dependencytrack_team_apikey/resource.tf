resource "dependencytrack_team" "example" {
  name = "Example"
}

resource "dependencytrack_team_apikey" "example" {
  team    = dependencytrack_team.id
  comment = "Example Comment"
}
