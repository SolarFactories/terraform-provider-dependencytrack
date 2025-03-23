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
					resource.TestCheckResourceAttr("dependencytrack_project.test", "version", ""),
					resource.TestCheckNoResourceAttr("dependencytrack_project.test", "parent"),
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
					resource.TestCheckResourceAttr("dependencytrack_project.test", "version", ""),
					resource.TestCheckNoResourceAttr("dependencytrack_project.test", "parent"),
				),
			},
		},
	})
}

func TestAccProjectNestedResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: providerConfig + `
resource "dependencytrack_project" "parent" {
	name = "Parent_Project"
}
resource "dependencytrack_project" "child" {
	name = "Child_Project"
	parent = dependencytrack_project.parent.id
}
`,
				Check: resource.TestCheckResourceAttrPair(
					"dependencytrack_project.parent", "id",
					"dependencytrack_project.child", "parent",
				),
			},
			// ImportState
			{
				ResourceName:      "dependencytrack_project.child",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read
			{
				Config: providerConfig + `
resource "dependencytrack_project" "parent" {
	name = "Parent_Project"
}
resource "dependencytrack_project" "child" {
	name = "Child_Project_With_Change"
	parent = dependencytrack_project.parent.id
}
`,
				Check: resource.TestCheckResourceAttrPair(
					"dependencytrack_project.parent", "id",
					"dependencytrack_project.child", "parent",
				),
			},
		},
	})
}
