# Requires DependencyTrack API v4.12+

resource "dependencytrack_project" "example" {
  name = "Example Project"
  tags = ["example_tag"]
}

resource "dependencytrack_notification_publisher" "example" {
  name               = "Example Publisher"
  publisher_class    = "org.dependencytrack.notification.publisher.ConsolePublisher"
  template_mime_type = "text/plain"
}

resource "dependencytrack_notification_rule" "example" {
  name                   = "Example Event Rule"
  trigger_type           = "EVENT"
  log_successful_publish = false
  notify_on = [
    "NEW_VULNERABILITY",
    "PROJECT_CREATED",
    "BOM_PROCESSED"
  ]
  publisher_id = dependencytrack_notification_publisher.test.id
}

resource "dependencytrack_tag_notification_rules" "example" {
  tag = "example_tag"
  policies = [
    dependencytrack_notification_rule.example.id,
  ]
  depends_on = [dependencytrack_project.example]
}
