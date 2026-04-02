package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationPublisherResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Notification_Publisher"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_notification_publisher.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_notification_publisher.test", "name", "Test_Notification_Publisher"),
					resource.TestCheckResourceAttr("dependencytrack_notification_publisher.test", "description", ""),
					resource.TestCheckResourceAttr("dependencytrack_notification_publisher.test", "publisher_class",
						"org.dependencytrack.notification.publisher.ConsolePublisher"),
					resource.TestCheckResourceAttr("dependencytrack_notification_publisher.test", "template", ""),
					resource.TestCheckResourceAttr("dependencytrack_notification_publisher.test", "template_mime_type", "text/plain"),
					resource.TestCheckResourceAttr("dependencytrack_notification_publisher.test", "default_publisher", "false"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_notification_publisher.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Notification_Publisher_With_Changes"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	description = "Test Description"
	template_mime_type = "text/plain"
	template = "Test Template"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_notification_publisher.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_notification_publisher.test", "name", "Test_Notification_Publisher_With_Changes"),
					resource.TestCheckResourceAttr("dependencytrack_notification_publisher.test", "description", "Test Description"),
					resource.TestCheckResourceAttr("dependencytrack_notification_publisher.test", "publisher_class",
						"org.dependencytrack.notification.publisher.ConsolePublisher"),
					resource.TestCheckResourceAttr("dependencytrack_notification_publisher.test", "template", "Test Template"),
					resource.TestCheckResourceAttr("dependencytrack_notification_publisher.test", "template_mime_type", "text/plain"),
					resource.TestCheckResourceAttr("dependencytrack_notification_publisher.test", "default_publisher", "false"),
				),
			},
		},
	})
}
