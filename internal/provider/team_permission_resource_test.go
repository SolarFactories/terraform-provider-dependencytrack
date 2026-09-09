package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccTeamPermissionResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_team_permission" "test" {
	team = dependencytrack_team.test.id
	permission = "SYSTEM_CONFIGURATION"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_team_permission.test", "team",
						"dependencytrack_team.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_team_permission.test", "permission", "SYSTEM_CONFIGURATION"),
				),
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_team_permission" "test" {
	team = dependencytrack_team.test.id
	permission = "BOM_UPLOAD"
}

`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_team_permission.test", "team",
						"dependencytrack_team.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_team_permission.test", "permission", "BOM_UPLOAD"),
				),
			},
		},
	})
}

// NOTE: Since `dependencytrack_team_permission` does not implement `Import`, cannot duplicate reference in state.
// Removing a Permission from a Team is a 200 OK response regardless of whether it was applied, in both API v4, v5.
// So `Delete` works seamlessly, manually verified after artificially inserting multiple `dependencytrack_team_permission`'s into tfstate against v5.1.0.
func TestAccTeamPermissionResourceRegression236(t *testing.T) {
	t.SkipNow()
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial Team Permission.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team_Permission_236"
}
resource "dependencytrack_team_permission" "test" {
	team = dependencytrack_team.test.id
	permission = "BOM_UPLOAD"
}
`,
			},
			// Duplicate reference to Team Permission.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team_Permission_236"
}
resource "dependencytrack_team_permission" "test" {
	team = dependencytrack_team.test.id
	permission = "BOM_UPLOAD"
}
resource "dependencytrack_team_permission" "test2" {
	team = dependencytrack_team.test.id
	permission = "BOM_UPLOAD"
}
import {
	to = dependencytrack_team_permission.test2
	id = dependencytrack_team_permission.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_team_permission.test", "id",
						"dependencytrack_team_permission.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team_Permission_236"
}
resource "dependencytrack_team_permission" "test2" {
	team = dependencytrack_team.test.id
	permission = "BOM_UPLOAD"
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_team_permission.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
