package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccProjectPropertyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_ProjectProperty"
}
resource "dependencytrack_project" "testencrypted" {
	name = "Test_ProjectPropertyEncrypted"
}
resource "dependencytrack_project_property" "test" {
	project = dependencytrack_project.test.id
	group = "A"
	name = "B"
	value = "C"
	type = "STRING"
	description = "D"
}
resource "dependencytrack_project_property" "testencrypted" {
	project = dependencytrack_project.testencrypted.id
	group = "G-Enc"
	name = "N-Enc"
	value = "TEST_ENCRYPTED_VALUE"
	type = "ENCRYPTEDSTRING"
	description = "D-Enc"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_project_property.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_project_property.test", "project",
						"dependencytrack_project.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_project_property.test", "group", "A"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.test", "name", "B"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.test", "value", "C"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.test", "type", "STRING"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.test", "description", "D"),
					//
					resource.TestCheckResourceAttrPair(
						"dependencytrack_project_property.testencrypted", "project",
						"dependencytrack_project.testencrypted", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "group", "G-Enc"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "name", "N-Enc"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "value", "TEST_ENCRYPTED_VALUE"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "type", "ENCRYPTEDSTRING"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "description", "D-Enc"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_project_property.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:            "dependencytrack_project_property.testencrypted",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_ProjectProperty"
}
resource "dependencytrack_project" "testencrypted" {
	name = "Test_ProjectPropertyEncrypted"
}
resource "dependencytrack_project_property" "test" {
	project = dependencytrack_project.test.id
	group = "A"
	name = "B"
	value = "2"
	type = "INTEGER"
	description = "D"
}
resource "dependencytrack_project_property" "testencrypted" {
	project = dependencytrack_project.testencrypted.id
	group = "G-Enc"
	name = "N-Enc"
	value = "TEST_ENCRYPTED_VALUE_WITH_CHANGE"
	type = "ENCRYPTEDSTRING"
	description = "D-Enc"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_project_property.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_project_property.test", "project",
						"dependencytrack_project.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_project_property.test", "group", "A"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.test", "name", "B"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.test", "value", "2"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.test", "type", "INTEGER"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.test", "description", "D"),
					//
					resource.TestCheckResourceAttrPair(
						"dependencytrack_project_property.testencrypted", "project",
						"dependencytrack_project.testencrypted", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "group", "G-Enc"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "name", "N-Enc"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "value", "TEST_ENCRYPTED_VALUE_WITH_CHANGE"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "type", "ENCRYPTEDSTRING"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "description", "D-Enc"),
				),
			},
		},
	})
}

func TestAccProjectPropertyResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial Project Property.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_ProjectProperty_236"
}
resource "dependencytrack_project_property" "test" {
	project = dependencytrack_project.test.id
	group = "A"
	name = "B"
	value = "2"
	type = "INTEGER"
}
`,
			},
			// Duplicate reference to Project Property.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_ProjectProperty_236"
}
resource "dependencytrack_project_property" "test" {
	project = dependencytrack_project.test.id
	group = "A"
	name = "B"
	value = "2"
	type = "INTEGER"
}
resource "dependencytrack_project_property" "test2" {
	project = dependencytrack_project.test.id
	group = "A"
	name = "B"
	value = "2"
	type = "INTEGER"
}
import {
	to = dependencytrack_project_property.test2
	id = dependencytrack_project_property.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_project_property.test", "id",
						"dependencytrack_project_property.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_ProjectProperty_236"
}
resource "dependencytrack_project_property" "test2" {
	project = dependencytrack_project.test.id
	group = "A"
	name = "B"
	value = "2"
	type = "INTEGER"
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_project_property.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
