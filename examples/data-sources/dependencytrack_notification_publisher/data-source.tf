data "dependencytrack_notification_publisher" "example" {
  name = "Slack"
}

output "dependencytrack_notification_publisher_id" {
  value = data.dependencytrack_notification_publisher.example.id
}
