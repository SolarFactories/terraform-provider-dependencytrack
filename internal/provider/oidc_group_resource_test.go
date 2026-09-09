package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccOidcGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_group" "test" {
	name = "Test_Group"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_oidc_group.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_oidc_group.test", "name", "Test_Group"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_oidc_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_group" "test" {
	name = "Test_Group_2"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_oidc_group.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_oidc_group.test", "name", "Test_Group_2"),
				),
			},
		},
	})
}

func TestAccOIDCGroupResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial Group.
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_group" "test" {
	name = "Test_OIDC_Group_236"
}
`,
			},
			// Duplicate reference to the Group.
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_group" "test" {
	name = "Test_OIDC_Group_236"
}
resource "dependencytrack_oidc_group" "test2" {
	name = "Test_OIDC_Group_236"
}
import {
	to = dependencytrack_oidc_group.test2
	id = dependencytrack_oidc_group.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_oidc_group.test", "id",
						"dependencytrack_oidc_group.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_group" "test2" {
	name = "Test_OIDC_Group_236"
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_oidc_group.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
