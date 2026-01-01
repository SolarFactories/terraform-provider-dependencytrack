package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccComponentsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_Components_Project"
}
resource "dependencytrack_component" "test" {
	project = dependencytrack_project.test.id
	name = "Test_Components_Component"
	version = "v1.0"
	classifier = "FILE"
	hashes = {}
}

data "dependencytrack_components" "test" {
	project = dependencytrack_project.test.id
	depends_on = [
		dependencytrack_component.test
	]
}

`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dependencytrack_components.test", "components.#", "1"),
					resource.TestCheckResourceAttrPair(
						"data.dependencytrack_components.test", "components.0.id",
						"dependencytrack_component.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.dependencytrack_components.test", "components.0.project",
						"dependencytrack_project.test", "id",
					),
				),
			},
		},
	})
}
