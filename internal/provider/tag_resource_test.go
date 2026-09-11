package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccTagResource(t *testing.T) {
	if apiSemver.Major < 4 || (apiSemver.Major == 4 && apiSemver.Minor < 13) {
		t.SkipNow()
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_tag" "test" {
	name = "test_tags_tag"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag.test", "id", "test_tags_tag"),
					resource.TestCheckResourceAttr("dependencytrack_tag.test", "name", "test_tags_tag"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_tag.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `

resource "dependencytrack_tag" "test" {
	name = "test_tags_tag_with_change"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag.test", "id", "test_tags_tag_with_change"),
					resource.TestCheckResourceAttr("dependencytrack_tag.test", "name", "test_tags_tag_with_change"),
				),
			},
		},
	})
}

func TestAccTagResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	if apiSemver.Major < 4 || (apiSemver.Major == 4 && apiSemver.Minor < 13) {
		t.SkipNow()
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial Tag.
			{
				Config: providerConfig + `
resource "dependencytrack_tag" "test" {
	name = "test_tag_236"
}
`,
			},
			// Duplicate reference to Tag.
			{
				Config: providerConfig + `
resource "dependencytrack_tag" "test" {
	name = "test_tag_236"
}
resource "dependencytrack_tag" "test2" {
	name = "test_tag_236"
}
import {
	to = dependencytrack_tag.test2
	id = dependencytrack_tag.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag.test", "id",
						"dependencytrack_tag.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_tag" "test2" {
	name = "test_tag_236"
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_tag.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
