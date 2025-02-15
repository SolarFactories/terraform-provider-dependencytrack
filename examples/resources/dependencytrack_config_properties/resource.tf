resource "dependencytrack_config_properties" "example" {
  properties = [
    {
      group = "general"
      name  = "base.url"
      value = "http://localhost:8000"
      type  = "STRING"
    },
    {
      group = "general"
      name  = "badge.enabled"
      value = "true"
      type  = "BOOLEAN"
    }
  ]
}
