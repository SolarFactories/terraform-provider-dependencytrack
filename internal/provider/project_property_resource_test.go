package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccProjectPropertyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_ProjectProperty"
	active = true
}
resource "dependencytrack_project_property" "test" {
	project = dependencytrack_project.test.id
	group = "A"
	name = "B"
	value = "C"
	type = "STRING"
	description = "D"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(),
			},
			// ImportState testing
			{
				ResourceName:      "dependencytrack_project_project.test",
				RefreshState:      true,
				ImportState:       false,
				ImportStateVerify: false,
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_ProjectProperty"
}
resource "dependencytrack_project_property" "test" {
	project = dependencytrack_project.test.id
	group = "A"
	name = "B"
	value = "C2"
	type = "STRING"
	description = "D"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_project_property.test", "group", "A"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.test", "name", "B"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.test", "value", "C2"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.test", "type", "STRING"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.test", "description", "D"),
				),
			},
		},
	})
}
