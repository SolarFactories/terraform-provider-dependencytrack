resource "dependencytrack_user" "example" {
  username = "Example"
  fullname = "Example User"
  email    = "Example_User@example.com"
  password = "Initial_User_Password"
}

resource "dependencytrack_user_permission" "example" {
  username   = dependencytrack_user.example.username
  permission = "BOM_UPLOAD"
}
