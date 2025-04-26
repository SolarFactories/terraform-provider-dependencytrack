resource "dependencytrack_team" "example" {
  name = "Example Team Name"
}

resource "dependencytrack_ldap_team_mapping" "example" {
  team               = dependencytrack_team.example.id
  distinguished_name = "example.ldap.server"
}
