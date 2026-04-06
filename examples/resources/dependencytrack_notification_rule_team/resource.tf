resource "dependencytrack_notification_publisher" "example" {
  name               = "Example Publisher"
  publisher_class    = "org.dependencytrack.notification.publisher.SendMailPublisher"
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

resource "dependencytrack_team" "example" {
  name = "Example Team"
}

resource "dependencytrack_notification_rule_team" "example" {
  rule = dependencytrack_notification_rule.example.id
  team = dependencytrack_team.example.id
}
