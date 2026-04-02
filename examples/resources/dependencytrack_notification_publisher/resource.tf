resource "dependencytrack_notification_publisher" "example" {
  name               = "Example"
  publisher_class    = "org.dependencytrack.notification.publisher.ConsolePublisher"
  description        = "Example Description"
  template_mime_type = "text/plain"
  template           = "Example Template"
}
