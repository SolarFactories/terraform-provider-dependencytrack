package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_Project"
	active = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_project.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "name", "Test_Project"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "active", "true"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "description", ""),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dependencytrack_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_Project_With_Change"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_project.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "name", "Test_Project_With_Change"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "active", "true"),
					resource.TestCheckResourceAttr("dependencytrack_project.test", "description", ""),
				),
			},
		},
	})
}
