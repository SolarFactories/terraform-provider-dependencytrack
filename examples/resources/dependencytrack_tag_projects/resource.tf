# Requires DependencyTrack API v4.12+

resource "dependencytrack_project" "example" {
  name = "Example Project"
  tags = ["example_tag"]
}

resource "dependencytrack_project" "example2" {
  name = "A Second Project"
}

resource "dependencytrack_tag_projects" "example" {
  tag = "example_tag"
  projects = [
    dependencytrack_project.example2.id,
    dependencytrack_project.example.id,
  ]
}
