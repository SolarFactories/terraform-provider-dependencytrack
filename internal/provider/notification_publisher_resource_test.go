package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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

func TestAccNotificationPublisherResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial Notification Publisher.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Notification_Publisher_236"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
	template = "Test"
}
`,
			},
			// Duplicate reference to Notification Publisher.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Notification_Publisher_236"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
	template = "Test"
}
resource "dependencytrack_notification_publisher" "test2" {
	name = "Test_Notification_Publisher_236"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
	template = "Test"
}
import {
	to = dependencytrack_notification_publisher.test2
	id = dependencytrack_notification_publisher.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_publisher.test", "id",
						"dependencytrack_notification_publisher.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test2" {
	name = "Test_Notification_Publisher_236"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_notification_publisher.test", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
