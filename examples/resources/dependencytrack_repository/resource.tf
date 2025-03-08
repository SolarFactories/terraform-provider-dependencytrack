resource "dependencytrack_repository" "example" {
  type       = "GITHUB"
  identifier = "github.com"
  url        = "https://github.com"
  enabled    = true
  username   = ""
  password   = ""
}
