package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccTeamPermissionsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_team_permissions" "test" {
	team = dependencytrack_team.test.id
	permissions = [
		"BOM_UPLOAD",
		"SYSTEM_CONFIGURATION",
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_team_permissions.test", "team",
						"dependencytrack_team.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_team_permissions.test", "permissions.#", "2"),
					resource.TestCheckResourceAttr("dependencytrack_team_permissions.test", "permissions.0", "BOM_UPLOAD"),
					resource.TestCheckResourceAttr("dependencytrack_team_permissions.test", "permissions.1", "SYSTEM_CONFIGURATION"),
				),
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_team_permissions" "test" {
	team = dependencytrack_team.test.id
	permissions = [
		"ACCESS_MANAGEMENT",
		"SYSTEM_CONFIGURATION",
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_team_permissions.test", "team",
						"dependencytrack_team.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_team_permissions.test", "permissions.#", "2"),
					resource.TestCheckResourceAttr("dependencytrack_team_permissions.test", "permissions.0", "ACCESS_MANAGEMENT"),
					resource.TestCheckResourceAttr("dependencytrack_team_permissions.test", "permissions.1", "SYSTEM_CONFIGURATION"),
				),
			},
		},
	})
}

func TestAccTeamPermissionsResourceRegression117(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/117
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_team_permissions" "test" {
	team = dependencytrack_team.test.id
	permissions = [
		"VIEW_PORTFOLIO",
		"BOM_UPLOAD",
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_team_permissions.test", "permissions.#", "2"),
					resource.TestCheckResourceAttr("dependencytrack_team_permissions.test", "permissions.0", "VIEW_PORTFOLIO"),
					resource.TestCheckResourceAttr("dependencytrack_team_permissions.test", "permissions.1", "BOM_UPLOAD"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_team_permissions.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_test" "test" {
	name = "Test_Team"
}
resource "dependencytrack_team_permissions" "test" {
	permissions = [
		"VIEW_PORTFOLIO",
		"BOM_UPLOAD",
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_team_permissions.test", "permissions.#", "2"),
					resource.TestCheckResourceAttr("dependencytrack_team_permissions.test", "permissions.0", "VIEW_PORTFOLIO"),
					resource.TestCheckResourceAttr("dependencytrack_team_permissions.test", "permissions.1", "BOM_UPLOAD"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
					PostApplyPreRefresh:  nil,
					PostApplyPostRefresh: nil,
				},
			},
		},
	})
}
