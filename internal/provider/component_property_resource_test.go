package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_component_property.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_component_property.test", "component",
						"dependencytrack_component.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "group", "A"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "name", "B"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "value", "C"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "type", "STRING"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "description", "D"),
				),
			},
			// ImportState testing.
			// TODO: Importing requires `id` and `component`. So would need to be custom constructed
			/*{
				ResourceName:      "dependencytrack_component_property.test",
				ImportState:       true,
				ImportStateVerify: true,
			},*/
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
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_component_property.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_component_property.test", "component",
						"dependencytrack_component.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "group", "A"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "name", "B"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "value", "2"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "type", "INTEGER"),
					resource.TestCheckResourceAttr("dependencytrack_component_property.test", "description", "D"),
				),
			},
		},
	})
}

func TestAccComponentPropertyResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial Component Property.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_ComponentProperty_236"
}
resource "dependencytrack_component" "test" {
	project = dependencytrack_project.test.id
	name = "Test_ComponentProperty_Component_236"
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
`,
			},
			// Duplicate reference to Component Property.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_ComponentProperty_236"
}
resource "dependencytrack_component" "test" {
	project = dependencytrack_project.test.id
	name = "Test_ComponentProperty_Component_236"
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
resource "dependencytrack_component_property" "test2" {
	component = dependencytrack_component.test.id
	group = "A"
	name = "B"
	value = "2"
	type = "INTEGER"
	description = "D"
}
import {
	to = dependencytrack_component_property.test2
	id = "${dependencytrack_component.test.id}/${dependencytrack_component_property.test.id}"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_component_property.test", "id",
						"dependencytrack_component_property.test2", "id",
					),
				),
			},
			// Remove one, causing other to be removed from state during Read, and handled absence for Delete.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_ComponentProperty_236"
}
resource "dependencytrack_component" "test" {
	project = dependencytrack_project.test.id
	name = "Test_ComponentProperty_Component_236"
	version = "v1.0"
	hashes = {}
}
resource "dependencytrack_component_property" "test2" {
	component = dependencytrack_component.test.id
	group = "A"
	name = "B"
	value = "2"
	type = "INTEGER"
	description = "D"
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_component_property.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
