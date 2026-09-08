package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccNotificationRuleEventResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Publisher_Event"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test" {
	name = "Test_Rule_Name_Event"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_notification_rule.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "name", "Test_Rule_Name_Event"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule.test", "publisher_id",
						"dependencytrack_notification_publisher.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "trigger_type", "EVENT"),
					// TODO: resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_children", "true"), API 4.12+.
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "log_successful_publish", "true"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "scope", "PORTFOLIO"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notification_level", "INFORMATIONAL"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_on.#", "0"),
				),
			},
			// ImportState testing.
			{
				ResourceName:            "dependencytrack_notification_rule.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"notify_children"}, // API 4.12+. Importing state in 4.13+ is covered by below test.
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Publisher_Event"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test" {
	name = "Test_Rule_Name_Event"
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
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "name", "Test_Rule_Name_Event"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule.test", "publisher_id",
						"dependencytrack_notification_publisher.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "trigger_type", "EVENT"),
					// TODO: resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_children", "true"), API 4.12+.
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "log_successful_publish", "false"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "scope", "PORTFOLIO"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notification_level", "INFORMATIONAL"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_on.#", "3"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_on.0", "NEW_VULNERABILITY"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_on.1", "PROJECT_CREATED"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_on.2", "BOM_PROCESSED"),
				),
			},
		},
	})
}

func TestAccNotificationRuleScheduleResource(t *testing.T) {
	// API 4.13+.
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Publisher_Schedule"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test" {
	name = "Test_Rule_Name_Schedule"
	trigger_type = "SCHEDULE"
	schedule_cron = "0 0 * * 0"
	publisher_id = dependencytrack_notification_publisher.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_notification_rule.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "name", "Test_Rule_Name_Schedule"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule.test", "publisher_id",
						"dependencytrack_notification_publisher.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "trigger_type", "SCHEDULE"),
					// TODO: resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_children", "true"), API 4.12+.
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "log_successful_publish", "true"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "scope", "PORTFOLIO"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notification_level", "INFORMATIONAL"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_on.#", "0"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "schedule_cron", "0 0 * * 0"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "schedule_skip_unchanged", "false"),
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
	name = "Test_Rule_Publisher_Schedule"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test" {
	name = "Test_Rule_Name_Schedule"
	trigger_type = "SCHEDULE"
	schedule_cron = "0 0 * * 1"
	schedule_skip_unchanged = true
	publisher_id = dependencytrack_notification_publisher.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_notification_rule.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "name", "Test_Rule_Name_Schedule"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule.test", "publisher_id",
						"dependencytrack_notification_publisher.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "trigger_type", "SCHEDULE"),
					// TODO: resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_children", "true"), API 4.12+.
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "log_successful_publish", "true"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "scope", "PORTFOLIO"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notification_level", "INFORMATIONAL"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "notify_on.#", "0"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "schedule_cron", "0 0 * * 1"),
					resource.TestCheckResourceAttr("dependencytrack_notification_rule.test", "schedule_skip_unchanged", "true"),
				),
			},
		},
	})
}

func TestAccNotificationRuleResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial Notification Rule.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Publisher_Event_236"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test" {
	name = "Test_Rule_Name_Event_236"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
`,
			},
			// Duplicate reference to Notification Rule.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Publisher_Event_236"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test" {
	name = "Test_Rule_Name_Event_236"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_notification_rule" "test2" {
	name = "Test_Rule_Name_Event_236"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
import {
	to = dependencytrack_notification_rule.test2
	id = dependencytrack_notification_rule.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_notification_rule.test", "id",
						"dependencytrack_notification_rule.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Publisher_Event_236"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test2" {
	name = "Test_Rule_Name_Event_236"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_notification_rule.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
