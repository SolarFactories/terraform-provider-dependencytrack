package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccLDAPMappingResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_ldap_team_mapping" "test" {
	team = dependencytrack_team.test.id
	distinguished_name = "test.mapping.ldap.dependencytrack.solarfactories"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_ldap_team_mapping.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_ldap_team_mapping.test", "distinguished_name", "test.mapping.ldap.dependencytrack.solarfactories"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_ldap_team_mapping.test", "team",
						"dependencytrack_team.test", "id",
					),
				),
			},
			// ImportState testing.
			{
				ImportState:       true,
				ResourceName:      "dependencytrack_ldap_team_mapping.test",
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					resources := state.RootModule().Resources
					teamID := resources["dependencytrack_ldap_team_mapping.test"].Primary.Attributes["team"]
					mappingID := resources["dependencytrack_ldap_team_mapping.test"].Primary.ID
					return teamID + "/" + mappingID, nil
				},
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_ldap_team_mapping" "test" {
	team = dependencytrack_team.test.id
	distinguished_name = "test2.mapping.ldap.dependencytrack.solarfactories"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_ldap_team_mapping.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_ldap_team_mapping.test", "distinguished_name", "test2.mapping.ldap.dependencytrack.solarfactories"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_ldap_team_mapping.test", "team",
						"dependencytrack_team.test", "id",
					),
				),
			},
		},
	})
}

func TestAccLDAPMappingResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial Mapping.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_LDAPMapping_236"
}
resource "dependencytrack_ldap_team_mapping" "test" {
	team = dependencytrack_team.test.id
	distinguished_name = "test236.mapping.ldap.dependencytrack.solarfactories"
}
`,
			},
			// Duplicate reference to the mapping.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_LDAPMapping_236"
}
resource "dependencytrack_ldap_team_mapping" "test" {
	team = dependencytrack_team.test.id
	distinguished_name = "test236.mapping.ldap.dependencytrack.solarfactories"
}
resource "dependencytrack_ldap_team_mapping" "test2" {
	team = dependencytrack_team.test.id
	distinguished_name = "test236.mapping.ldap.dependencytrack.solarfactories"
}
import {
	to = dependencytrack_ldap_team_mapping.test2
	id = "${dependencytrack_team.test.id}/${dependencytrack_ldap_team_mapping.test.id}"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_ldap_team_mapping.test", "id",
						"dependencytrack_ldap_team_mapping.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_LDAPMapping_236"
}
resource "dependencytrack_ldap_team_mapping" "test2" {
	team = dependencytrack_team.test.id
	distinguished_name = "test236.mapping.ldap.dependencytrack.solarfactories"
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_ldap_team_mapping.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
