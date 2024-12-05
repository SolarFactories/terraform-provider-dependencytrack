package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccTeamResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Project"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_team.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_team.test", "name", "Test_Project"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dependencytrack_team.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Project"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_team.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_team.test", "name", "Test_Project"),
				),
			},
		},
	})
}
