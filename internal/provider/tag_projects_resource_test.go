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
resource "dependencytrack_project" "test" {
	name = "TagProjectsProject"
	tags = ["test_projects_tag"]
}
resource "dependencytrack_project" "test2" {
	name = "TestProjectsProject2"
	version = "v1"
}
resource "dependencytrack_tag_projects" "test" {
	tag = "test_projects_tag"
	projects = [
		dependencytrack_project.test.id,
		dependencytrack_project.test2.id,
	]
}
data "dependencytrack_project" "test2" {
	name = dependencytrack_project.test2.name
	version = dependencytrack_project.test2.version
	depends_on = [dependencytrack_tag_projects.test]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag_projects.test", "projects.#", "2"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_projects.test", "projects.0",
						"dependencytrack_project.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_projects.test", "projects.1",
						"dependencytrack_project.test2", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_tag_projects.test", "tag", "test_projects_tag"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test2", "tags.#", "1"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test2", "tags.0", "test_projects_tag"),
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
resource "dependencytrack_project" "test" {
	name = "TagProjectsProject"
	tags = ["test_projects_tag"]
}
resource "dependencytrack_project" "test2" {
	name = "TestProjectsProject2"
}
resource "dependencytrack_tag_projects" "test" {
	tag = "test_projects_tag"
	projects = [
		dependencytrack_project.test.id,
	]
}
data "dependencytrack_project" "test2" {
	name = dependencytrack_project.test2.name
	version = dependencytrack_project.test2.version
	depends_on = [dependencytrack_tag_projects.test]
}

`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag_projects.test", "projects.#", "1"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_tag_projects.test", "projects.0",
						"dependencytrack_project.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_tag_projects.test", "tag", "test_projects_tag"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test2", "tags.#", "0"),
				),
			},
		},
	})
}
