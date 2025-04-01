package provider

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccProjectPropertyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_ProjectProperty"
	active = true
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
	project = dependencytrack_project.test.id
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
						"dependencytrack_project.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "group", "G-Enc"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "name", "N-Enc"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "value", "TEST_ENCRYPTED_VALUE"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "type", "ENCRYPTEDSTRING"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "description", "D-Enc"),
				),
			},
			// ImportState testing
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
			// Update and Read testing
			{
				Config: providerConfig + `
resource "dependencytrack_project" "test" {
	name = "Test_ProjectProperty"
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
	project = dependencytrack_project.test.id
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
						"dependencytrack_project.test", "id",
					),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "group", "G-Enc"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "name", "N-Enc"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "value", "TEST_ENCRYPTED_VALUE_WITH_CHANGE"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "type", "ENCRYPTEDSTRING"),
					resource.TestCheckResourceAttr("dependencytrack_project_property.testencrypted", "description", "D-Enc"),
				),
			},
			// Sleep, to debug tests, before destroying ProjectProperties
			{
				Destroy:      true,
				ResourceName: "dependencytrack_project_property.testencrypted",
				RefreshState: true,
			},
			{
				Destroy:      true,
				ResourceName: "dependencytrack_project_property.test",
				RefreshState: true,
			},
			{
				RefreshState: true,
				Check: func(s *terraform.State) error {
					duration, err := time.ParseDuration("10s")
					if err != nil {
						return err
					}
					time.Sleep(duration)
					return nil
				},
			},
		},
	})
}
