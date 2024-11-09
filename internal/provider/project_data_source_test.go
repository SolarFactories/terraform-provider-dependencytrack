package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccProjectDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "dependencytrack_project" "test" {
	name = "Project_Data_Test"
	version = "1"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dependencytrack_project.test", "id"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "name", "Project_Data_Test"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "version", "1"),
					//
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "properties.#", "2"),
					//
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "properties.0.group", "Group1"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "properties.0.name", "Name1"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "properties.0.value", "Value1"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "properties.0.type", "STRING"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "properties.0.description", "Description1"),
					//
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "properties.1.group", "Group2"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "properties.1.name", "Name2"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "properties.1.value", "2"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "properties.1.type", "INTEGER"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "properties.1.description", "Description2"),
				),
			},
		},
	})
}
