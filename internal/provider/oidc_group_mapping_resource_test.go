package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccOidcGroupMappingResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_oidc_group" "test" {
	name = "Test_Group"
}
resource "dependencytrack_oidc_group_mapping" "test" {
	team = dependencytrack_team.test.id
	group = dependencytrack_oidc_group.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_oidc_group_mapping.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_oidc_group_mapping.test", "team",
						"dependencytrack_team.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_oidc_group_mapping.test", "group",
						"dependencytrack_oidc_group.test", "id",
					),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_oidc_group_mapping.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_oidc_group" "test" {
	name = "Test_Group"
}
resource "dependencytrack_oidc_group" "test2" {
	name = "Test_Group_2"
}
resource "dependencytrack_oidc_group_mapping" "test" {
	team = dependencytrack_team.test.id
	group = dependencytrack_oidc_group.test2.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_oidc_group_mapping.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_oidc_group_mapping.test", "team",
						"dependencytrack_team.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_oidc_group_mapping.test", "group",
						"dependencytrack_oidc_group.test2", "id",
					),
				),
			},
		},
	})
}

func TestAccOIDCGroupMappingResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial OIDC Group Mapping.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_OIDCGroupMapping_236"
}
resource "dependencytrack_oidc_group" "test" {
	name = "Test_OIDCGroupMapping_236"
}
resource "dependencytrack_oidc_group_mapping" "test" {
	team = dependencytrack_team.test.id
	group = dependencytrack_oidc_group.test.id
}
`,
			},
			// Duplicate reference to OIDC Group Mapping.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_OIDCGroupMapping_236"
}
resource "dependencytrack_oidc_group" "test" {
	name = "Test_OIDCGroupMapping_236"
}
resource "dependencytrack_oidc_group_mapping" "test" {
	team = dependencytrack_team.test.id
	group = dependencytrack_oidc_group.test.id
}
resource "dependencytrack_oidc_group_mapping" "test2" {
	team = dependencytrack_team.test.id
	group = dependencytrack_oidc_group.test.id
}
import {
	to = dependencytrack_oidc_group_mapping.test2
	id = dependencytrack_oidc_group_mapping.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_oidc_group_mapping.test", "id",
						"dependencytrack_oidc_group_mapping.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_OIDCGroupMapping_236"
}
resource "dependencytrack_oidc_group" "test" {
	name = "Test_OIDCGroupMapping_236"
}
resource "dependencytrack_oidc_group_mapping" "test2" {
	team = dependencytrack_team.test.id
	group = dependencytrack_oidc_group.test.id
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_oidc_group_mapping.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
