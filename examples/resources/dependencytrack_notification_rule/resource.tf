resource "dependencytrack_notification_publisher" "example" {
  name               = "Example Publisher"
  publisher_class    = "org.dependencytrack.notification.publisher.ConsolePublisher"
  template_mime_type = "text/plain"
}

// Event Driven.
resource "dependencytrack_notification_rule" "example_event" {
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

// Scheduled. Available in DependencyTrack API 4.13+
resource "dependencytrack_notification_rule" "example_schedule" {
  name          = "Example Schedule Rule"
  trigger_type  = "SCHEDULE"
  schedule_cron = "0 0 * * 0"
  publisher_id  = dependencytrack_notification_publisher.test.id
}

