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

resource "dependencytrack_project" "example" {
  name = "Example Project"
}

resource "dependencytrack_notification_rule_project" "example" {
  rule    = dependencytrack_notification_rule.example.id
  project = dependencytrack_project.example.id
}
