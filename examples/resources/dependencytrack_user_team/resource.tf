resource "dependencytrack_team" "example" {
  name = "Example"
}

resource "dependencytrack_user" "example" {
  username = "Example"
  fullname = "Example User"
  email    = "Example_User@example.com"
  password = "Initial_User_Password"
}

resource "dependencytrack_user_team" "example" {
  username = dependencytrack_user.example.username
  team     = dependencytrack_team.example.id
}
