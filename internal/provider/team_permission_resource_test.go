package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
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
