resource "dependencytrack_config_property" "example" {
  group = "general"
  name  = "base.url"
  value = "http://localhost:8000"
  type  = "STRING"
}
