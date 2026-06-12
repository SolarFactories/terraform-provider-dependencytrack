package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationPublisherDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
  name               = "Data Source Test Publisher"
  description        = "Publisher for testing data source"
  publisher_class    = "org.dependencytrack.notification.publisher.SlackPublisher"
  template_mime_type = "application/json"
}

data "dependencytrack_notification_publisher" "test" {
  name = dependencytrack_notification_publisher.test.name
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.dependencytrack_notification_publisher.test", "id",
						"dependencytrack_notification_publisher.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.dependencytrack_notification_publisher.test", "description",
						"dependencytrack_notification_publisher.test", "description",
					),
					resource.TestCheckResourceAttrPair(
						"data.dependencytrack_notification_publisher.test", "publisher_class",
						"dependencytrack_notification_publisher.test", "publisher_class",
					),
					resource.TestCheckResourceAttrPair(
						"data.dependencytrack_notification_publisher.test", "template_mime_type",
						"dependencytrack_notification_publisher.test", "template_mime_type",
					),
				),
			},
		},
	})
}
