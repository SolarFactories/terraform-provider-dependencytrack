package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTagNotificationRulesResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				// Use Project to create the Tag, due to `dependencytrack_tag` being 4.13+.
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Tag_Notification_Rules_Project"
	tags = ["tag_notification_rules"]
}
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Publisher_Tag"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test_a" {
	name = "Test_Rule_Publisher_Tag_A"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_notification_rule" "test_z" {
	name = "Test_Rule_Publisher_Tag_Z"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_tag_notification_rules" "test" {
	tag = "tag_notification_rules"
	notification_rules = [
		dependencytrack_notification_rule.test_a.id,
		dependencytrack_notification_rule.test_z.id,
	]
	depends_on = [dependencytrack_project.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag_notification_rules.test", "id", "tag_notification_rules"),
					resource.TestCheckResourceAttr("dependencytrack_tag_notification_rules.test", "notification_rules.#", "2"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_notification_rules.test", "notification_rules.0",
						"dependencytrack_notification_rule.test_a", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_notification_rules.test", "notification_rules.1",
						"dependencytrack_notification_rule.test_z", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_tag_notification_rules.test", "tag", "tag_notification_rules"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_tag_notification_rules.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Tag_Notification_Rules_Project"
	tags = ["tag_notification_rules"]
}
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Publisher_Tag"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test_a" {
	name = "Test_Rule_Publisher_Tag_A"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_notification_rule" "test_b" {
	name = "Test_Rule_Publisher_Tag_B"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_notification_rule" "test_z" {
	name = "Test_Rule_Publisher_Tag_Z"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_tag_notification_rules" "test" {
	tag = "tag_notification_rules"
	notification_rules = [
		dependencytrack_notification_rule.test_a.id,
		dependencytrack_notification_rule.test_b.id,
		dependencytrack_notification_rule.test_z.id,
	]
	depends_on = [dependencytrack_project.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag_notification_rules.test", "id", "tag_notification_rules"),
					resource.TestCheckResourceAttr("dependencytrack_tag_notification_rules.test", "notification_rules.#", "3"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_notification_rules.test", "notification_rules.0",
						"dependencytrack_notification_rule.test_a", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_notification_rules.test", "notification_rules.1",
						"dependencytrack_notification_rule.test_b", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_notification_rules.test", "notification_rules.2",
						"dependencytrack_notification_rule.test_z", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_tag_notification_rules.test", "tag", "tag_notification_rules"),
				),
			},
		},
	})
}

func TestAccTagNotificationRulesResourceNotificationRulesUnordered(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				// Use Project to create the Tag, due to `dependencytrack_tag` being 4.13+.
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Tag_Notification_Rules_Project_Unordered"
	tags = ["tag_notification_rules_unordered"]
}
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Publisher_Tag_Unordered"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test_a" {
	name = "A_Test_Rule_Publisher_Tag_Unordered"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_notification_rule" "test_z" {
	name = "Z_Test_Rule_Publisher_Tag_Unordered"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_tag_notification_rules" "test" {
	tag = "tag_notification_rules"
	notification_rules = [
		dependencytrack_notification_rule.test_z.id,
		dependencytrack_notification_rule.test_a.id,
	]
	depends_on = [dependencytrack_project.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag_notification_rules.test", "notification_rules.#", "2"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_notification_rules.test", "notification_rules.0",
						"dependencytrack_notification_rule.test_z", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_notification_rules.test", "notification_rules.1",
						"dependencytrack_notification_rule.test_a", "id",
					),
				),
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Tag_Notification_Rules_Project_Unordered"
	tags = ["tag_notification_rules_unordered"]
}
resource "dependencytrack_notification_publisher" "test" {
	name = "Test_Rule_Publisher_Tag_Unordered"
	publisher_class = "org.dependencytrack.notification.publisher.ConsolePublisher"
	template_mime_type = "text/plain"
}
resource "dependencytrack_notification_rule" "test_a" {
	name = "A_Test_Rule_Publisher_Tag_Unordered"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_notification_rule" "test_b" {
	name = "B_Test_Rule_Publisher_Tag_Unordered"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_notification_rule" "test_z" {
	name = "Z_Test_Rule_Publisher_Tag_Unordered"
	trigger_type = "EVENT"
	publisher_id = dependencytrack_notification_publisher.test.id
}
resource "dependencytrack_tag_notification_rules" "test" {
	tag = "tag_notification_rules"
	notification_rules = [
		dependencytrack_notification_rule.test_z.id,
		dependencytrack_notification_rule.test_a.id,
		dependencytrack_notification_rule.test_b.id,
	]
	depends_on = [dependencytrack_project.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag_notification_rules.test", "notification_rules.#", "3"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_notification_rules.test", "notification_rules.0",
						"dependencytrack_notification_rule.test_z", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_notification_rules.test", "notification_rules.1",
						"dependencytrack_notification_rule.test_a", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_notification_rules.test", "notification_rules.2",
						"dependencytrack_notification_rule.test_b", "id",
					),
				),
			},
		},
	})
}
