package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccComponentResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_Component_Project"
}
resource "dependencytrack_component" "test" {
	project = dependencytrack_project.test.id
	name = "Test_Component_Component"
	version = "v1.0"
	hashes = {
		md5 = "00000000000000000000000000000001"
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_component.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_project.test", "id",
						"dependencytrack_component.test", "project",
					),
					resource.TestCheckResourceAttr("dependencytrack_component.test", "name", "Test_Component_Component"),
					resource.TestCheckResourceAttr("dependencytrack_component.test", "version", "v1.0"),
					resource.TestCheckResourceAttr("dependencytrack_component.test", "classifier", "LIBRARY"),
					resource.TestCheckResourceAttr("dependencytrack_component.test", "hashes.%", "12"),
					resource.TestCheckResourceAttr("dependencytrack_component.test", "hashes.md5", "00000000000000000000000000000001"),
					resource.TestCheckResourceAttr("dependencytrack_component.test", "hashes.sha1", ""),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_component.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_Component_Project"
}
resource "dependencytrack_component" "test" {
	project = dependencytrack_project.test.id
	name = "Test_Component_Component"
	version = "v1.0"
	classifier = "FILE"
	hashes = {
		md5 = ""
	}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_component.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_project.test", "id",
						"dependencytrack_component.test", "project",
					),
					resource.TestCheckResourceAttr("dependencytrack_component.test", "name", "Test_Component_Component"),
					resource.TestCheckResourceAttr("dependencytrack_component.test", "version", "v1.0"),
					resource.TestCheckResourceAttr("dependencytrack_component.test", "classifier", "FILE"),
					resource.TestCheckResourceAttr("dependencytrack_component.test", "hashes.%", "12"),
					resource.TestCheckResourceAttr("dependencytrack_component.test", "hashes.md5", ""),
					resource.TestCheckResourceAttr("dependencytrack_component.test", "hashes.sha1", ""),
				),
			},
		},
	})
}

func TestAccComponentResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial Component.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_Component_236"
}
resource "dependencytrack_component" "test" {
	project = dependencytrack_project.test.id
	name = "Test_Component_236"
	version = "v1.0"
	hashes = {}
}
`,
			},
			// Duplicate reference to the component.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_Component_236"
}
resource "dependencytrack_component" "test" {
	project = dependencytrack_project.test.id
	name = "Test_Component_236"
	version = "v1.0"
	hashes = {}
}
resource "dependencytrack_component" "test2" {
	project = dependencytrack_project.test.id
	name = "Test_Component_236"
	version = "v1.0"
	hashes = {}
}
import {
	to = dependencytrack_component.test2
	id = dependencytrack_component.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_component.test", "id",
						"dependencytrack_component.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_Component_236"
}
resource "dependencytrack_component" "test2" {
	project = dependencytrack_project.test.id
	name = "Test_Component_236"
	version = "v1.0"
	hashes = {}
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_component.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
