package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTagProjectsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
data "dependencytrack_project" "test" {
	name = "Project_Data_Test"
	version = "1"
}
resource "dependencytrack_project" "test" {
	name = "TagProjectsProject"
	version = "1.0.0"
}
resource "dependencytrack_tag_projects" "test" {
	tag = "testtag"
	projects = [
		data.dependencytrack_project.test.id,
		dependencytrack_project.test.id,
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag_projects.test", "projects.#", "2"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_projects.test", "projects.0",
						"data.dependencytrack_project.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_projects.test", "projects.1",
						"dependencytrack_project.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_tag_projects.test", "tag", "testtag"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_tag_projects.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
data "dependencytrack_project" "test" {
	name = "Project_Data_Test"
	version = "1"
}
resource "dependencytrack_project" "test" {
	name = "TagProjectsProject"
	version = "1.0.0"
}
resource "dependencytrack_tag_projects" "test" {
	tag = "testtag"
	projects = [
		data.dependencytrack_project.test.id,
		dependencytrack_project.test.id,
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag_projects.test", "projects.#", "2"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_projects.test", "projects.0",
						"data.dependencytrack_project.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_projects.test", "projects.1",
						"dependencytrack_project.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_tag_projects.test", "tag", "testtag"),
				),
			},
		},
	})
}
