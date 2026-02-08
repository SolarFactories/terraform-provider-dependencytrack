package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "classifier", "APPLICATION"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "cpe", ""),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "group", ""),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "purl", ""),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "swid", ""),
					resource.TestCheckNoResourceAttr("data.dependencytrack_project.test", "parent"),
					//
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "tags.#", "1"),
					resource.TestCheckResourceAttr("data.dependencytrack_project.test", "tags.0", "project_data_test_tag"),
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

// API 4.12+.
func TestAccProjectDataSourceIsLatest(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Project_Data_Test_Latest"
	version = "1"
	is_latest = true
}
data "dependencytrack_project" "test2" {
	name = dependencytrack_project.test.name
	version = dependencytrack_project.test.version
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.dependencytrack_project.test2", "is_latest", "dependencytrack_project.test", "is_latest"),
				),
			},
		},
	})
}
