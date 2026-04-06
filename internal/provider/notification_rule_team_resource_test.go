package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationRuleTeamResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Team_Publisher"
	publisher_class = "org.dependencytrack.notification.publisher.SendMailPublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test" {
	name = "Test_Rule_Team_Name"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_team" "test" {
	name = "Test_Rule_Team"
}
resource "dependencytrack_notification_rule_team" "test" {
	rule = dependencytrack_notification_rule.test.id
	team = dependencytrack_team.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_notification_rule_team.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule_team.test", "rule",
						"dependencytrack_notification_rule.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule_team.test", "team",
						"dependencytrack_team.test", "id",
					),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_notification_rule_team.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Team_Publisher"
	publisher_class = "org.dependencytrack.notification.publisher.SendMailPublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test" {
	name = "Test_Rule_Team_Name"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_team" "test" {
	name = "Test_Rule_Team"
}
resource "dependencytrack_notification_rule_team" "test" {
	rule = dependencytrack_notification_rule.test.id
	team = dependencytrack_team.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_notification_rule_team.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule_team.test", "rule",
						"dependencytrack_notification_rule.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule_team.test", "team",
						"dependencytrack_team.test", "id",
					),
				),
			},
		},
	})
}
