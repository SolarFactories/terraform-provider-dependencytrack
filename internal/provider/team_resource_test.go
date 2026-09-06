package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTeamResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_team.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_team.test", "name", "Test_Team"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_team.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team_2"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_team.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_team.test", "name", "Test_Team_2"),
				),
			},
		},
	})
}

func TestAccTeamResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial team.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team_236"
}
`,
			},
			// Duplicate reference to team.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team_236"
}
resource "dependencytrack_team" "test2" {
	name = "Test_Team_236"
}
import {
	to = dependencytrack_team.test2
	id = dependencytrack_team.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_team.test", "id",
						"dependencytrack_team.test2", "id",
					),
				),
			},
			// Delete initial team, and read copied reference.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test2" {
	name = "Test_Team_236"
}
`,
			},
		},
	})
}
