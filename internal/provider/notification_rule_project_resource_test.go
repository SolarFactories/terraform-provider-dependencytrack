package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationRuleProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Project_Publisher"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test" {
	name = "Test_Rule_Project_Name"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_project" "test" {
	name = "Test_Rule_Project"
}
resource "dependencytrack_notification_rule_project" "test" {
	rule = dependencytrack_notification_rule.test.id
	project = dependencytrack_project.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_notification_rule_project.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule_project.test", "rule",
						"dependencytrack_notification_rule.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule_project.test", "project",
						"dependencytrack_project.test", "id",
					),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_notification_rule_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Project_Publisher"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test" {
	name = "Test_Rule_Project_Name"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_project" "test" {
	name = "Test_Rule_Project"
}
resource "dependencytrack_notification_rule_project" "test" {
	rule = dependencytrack_notification_rule.test.id
	project = dependencytrack_project.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_notification_rule_project.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule_project.test", "rule",
						"dependencytrack_notification_rule.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule_project.test", "project",
						"dependencytrack_project.test", "id",
					),
				),
			},
		},
	})
}
