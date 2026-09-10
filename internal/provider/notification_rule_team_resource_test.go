package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccNotificationRuleTeamResource(t *testing.T) {
	if apiSemver.Major > 4 {
		t.Skip("TODO: Notification Publisher API schema changed in v5.")
	}
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
				// NOTE: Can check to expect no change to plan.
			},
		},
	})
}

func TestAccNotificationRuleTeamResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial Notification Rule Team.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Team_Publisher_236"
	publisher_class = "org.dependencytrack.notification.publisher.SendMailPublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test" {
	name = "Test_Rule_Team_Name_236"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_team" "test" {
	name = "Test_Rule_Team_236"
}
resource "dependencytrack_notification_rule_team" "test" {
	rule = dependencytrack_notification_rule.test.id
	team = dependencytrack_team.test.id
}
`,
			},
			// Duplicate reference to Notification Rule Team.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Team_Publisher_236"
	publisher_class = "org.dependencytrack.notification.publisher.SendMailPublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test" {
	name = "Test_Rule_Team_Name_236"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_team" "test" {
	name = "Test_Rule_Team_236"
}
resource "dependencytrack_notification_rule_team" "test" {
	rule = dependencytrack_notification_rule.test.id
	team = dependencytrack_team.test.id
}
resource "dependencytrack_notification_rule_team" "test2" {
	rule = dependencytrack_notification_rule.test.id
	team = dependencytrack_team.test.id
}
import {
	to = dependencytrack_notification_rule_team.test2
	id = dependencytrack_notification_rule_team.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule_team.test", "id",
						"dependencytrack_notification_rule_team.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Team_Publisher_236"
	publisher_class = "org.dependencytrack.notification.publisher.SendMailPublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test" {
	name = "Test_Rule_Team_Name_236"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_team" "test" {
	name = "Test_Rule_Team_236"
}
resource "dependencytrack_notification_rule_team" "test2" {
	rule = dependencytrack_notification_rule.test.id
	team = dependencytrack_team.test.id
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_notification_rule_team.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
