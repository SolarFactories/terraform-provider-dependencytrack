package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationRuleEventResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Publisher"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test" {
	name = "Test_Rule_Name"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_notification_rule.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "name", "Test_Rule_Name"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule.test", "publisher_id",
						"dependencytrack_notification_publisher.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "enabled", "true"),
					// TODO: resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_children", "true"), API 4.12+.
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "log_successful_publish", "true"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "scope", "PORTFOLIO"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notification_level", ""),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_on.#", "0"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_notification_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Publisher"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test" {
	name = "Test_Rule_Name"
	trigger_type = "EVENT"
	log_successful_publish = false
	notify_on = [
		"NEW_VULNERABILITY",
		"PROJECT_CREATED",
		"BOM_PROCESSED"
	]
	publisher_id = dependencytrack_notification_publisher.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_notification_rule.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "name", "Test_Rule_Name"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule.test", "publisher_id",
						"dependencytrack_notification_publisher.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "enabled", "true"),
					// TODO: resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_children", "true"), API 4.12+.
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "log_successful_publish", "false"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "scope", "PORTFOLIO"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notification_level", ""),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_on.#", "3"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_on.0", "NEW_VULNERABILITY"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_on.1", "PROJECT_CREATED"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_on.2", "BOM_PROCESSED"),
				),
			},
		},
	})
}

/*func TestAccNotificationRuleScheduleResource(t *testing.T) {
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
}*/
