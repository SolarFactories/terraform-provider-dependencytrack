package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccComponentPropertyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_ComponentProperty"
}
resource "dependencytrack_component" "test" {
	project = dependencytrack_project.test.id
	name = "Test_ComponentProperty_Component"
	version = "v1.0"
	hashes = {}
}
resource "dependencytrack_component_property" "test" {
	component = dependencytrack_component.test.id
	group = "A"
	name = "B"
	value = "C"
	type = "STRING"
	description = "D"
}
resource "dependencytrack_component_property" "testencrypted" {
	component = dependencytrack_component.test.id
	group = "G-Enc"
	name = "N-Enc"
	value = "TEST_ENCRYPTED_VALUE"
	type = "ENCRYPTEDSTRING"
	description = "D-Enc"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_component_property.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_component_property.test", "project",
						"dependencytrack_component.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "group", "A"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "name", "B"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "value", "C"),
					resource.TestCheckResourceAttr("dependencytrack_compoment_property.test", "type", "STRING"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "description", "D"),
					//
					resource.TestCheckResourceAttrPair(
						"dependencytrack_component_property.testencrypted", "project",
						"dependencytrack_component.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_component_property.testencrypted", "group", "G-Enc"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.testencrypted", "name", "N-Enc"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.testencrypted", "value", "TEST_ENCRYPTED_VALUE"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.testencrypted", "type", "ENCRYPTEDSTRING"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.testencrypted", "description", "D-Enc"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_component_property.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:            "dependencytrack_component_property.testencrypted",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_ComponentProperty"
}
resource "dependencytrack_component" "test" {
	project = dependencytrack_project.test.id
	name = "Test_ComponentProperty_Component"
	version = "v1.0"
	hashes = {}
}
resource "dependencytrack_component_property" "test" {
	component = dependencytrack_component.test.id
	group = "A"
	name = "B"
	value = "2"
	type = "INTEGER"
	description = "D"
}
resource "dependencytrack_component_property" "testencrypted" {
	component = dependencytrack_component.test.id
	group = "G-Enc"
	name = "N-Enc"
	value = "TEST_ENCRYPTED_VALUE_WITH_CHANGE"
	type = "ENCRYPTEDSTRING"
	description = "D-Enc"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_component_property.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_component_property.test", "project",
						"dependencytrack_component.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "group", "A"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "name", "B"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "value", "2"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "type", "INTEGER"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "description", "D"),
					//
					resource.TestCheckResourceAttrPair(
						"dependencytrack_component_property.testencrypted", "project",
						"dependencytrack_component.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_component_property.testencrypted", "group", "G-Enc"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.testencrypted", "name", "N-Enc"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.testencrypted", "value", "TEST_ENCRYPTED_VALUE_WITH_CHANGE"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.testencrypted", "type", "ENCRYPTEDSTRING"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.testencrypted", "description", "D-Enc"),
				),
			},
		},
	})
}
