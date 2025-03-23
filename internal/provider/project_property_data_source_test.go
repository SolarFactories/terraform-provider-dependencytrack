package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccProjectPropertyDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "dependencytrack_project" "test" {
	name = "Project_Data_Test"
	version = "1"
}
data "dependencytrack_project_property" "test0" {
	project = data.dependencytrack_project.test.id
	group = "Group1"
	name = "Name1"
}
data "dependencytrack_project_property" "test1" {
	project = data.dependencytrack_project.test.id
	group = "Group2"
	name = "Name2"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.dependencytrack_project_property.test0", "project",
						"data.dependencytrack_project.test", "id",
					),
					resource.TestCheckResourceAttr("data.dependencytrack_project_property.test0", "group", "Group1"),
					resource.TestCheckResourceAttr("data.dependencytrack_project_property.test0", "name", "Name1"),
					resource.TestCheckResourceAttr("data.dependencytrack_project_property.test0", "value", "Value1"),
					resource.TestCheckResourceAttr("data.dependencytrack_project_property.test0", "type", "STRING"),
					resource.TestCheckResourceAttr("data.dependencytrack_project_property.test0", "description", "Description1"),
					//
					resource.TestCheckResourceAttrPair(
						"data.dependencytrack_project_property.test1", "project",
						"data.dependencytrack_project.test", "id",
					),
					resource.TestCheckResourceAttr("data.dependencytrack_project_property.test1", "group", "Group2"),
					resource.TestCheckResourceAttr("data.dependencytrack_project_property.test1", "name", "Name2"),
					resource.TestCheckResourceAttr("data.dependencytrack_project_property.test1", "value", "2"),
					resource.TestCheckResourceAttr("data.dependencytrack_project_property.test1", "type", "INTEGER"),
					resource.TestCheckResourceAttr("data.dependencytrack_project_property.test1", "description", "Description2"),
				),
			},
		},
	})
}
