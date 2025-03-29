package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccConfigPropertyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "dependencytrack_config_property" "testbool" {
	group = "email"
	name = "smtp.enabled"
	value = "true"
	type = "BOOLEAN"
}
resource "dependencytrack_config_property" "teststring" {
	group = "email"
	name = "subject.prefix"
	value = "TF Test"
	type = "STRING"
}
resource "dependencytrack_config_property" "testencrypted" {
	group = "email"
	name = "smtp.password"
	value = "TEST_PASSWORD"
	type = "ENCRYPTEDSTRING"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_config_property.testbool", "id", "email/smtp.enabled"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testbool", "group", "email"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testbool", "name", "smtp.enabled"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testbool", "value", "true"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testbool", "type", "BOOLEAN"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testbool", "description", "Flag to enable/disable SMTP"),
					//
					resource.TestCheckResourceAttr("dependencytrack_config_property.teststring", "id", "email/subject.prefix"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.teststring", "group", "email"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.teststring", "name", "subject.prefix"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.teststring", "value", "TF Test"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.teststring", "type", "STRING"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.teststring", "description", "The Prefix Subject email to use"),
					//
					resource.TestCheckResourceAttr("dependencytrack_config_property.testencrypted", "group", "email"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testencrypted", "name", "smtp.password"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testencrypted", "value", "TEST_PASSWORD"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testencrypted", "type", "ENCRYPTEDSTRING"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testencrypted", "description", "The optional password for the username used for authentication"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dependencytrack_config_property.testbool",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "dependencytrack_config_property.teststring",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:            "dependencytrack_config_property.testencrypted",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "dependencytrack_config_property" "testbool" {
	group = "email"
	name = "smtp.enabled"
	value = "false"
	type = "BOOLEAN"

}
resource "dependencytrack_config_property" "teststring" {
	group = "email"
	name = "subject.prefix"
	value = "TF Test with Update"
	type = "STRING"
}
resource "dependencytrack_config_property" "testencrypted" {
	group = "email"
	name = "smtp.password"
	value = "TEST_PASSWORD"
	type = "ENCRYPTEDSTRING"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_config_property.testbool", "id", "email/smtp.enabled"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testbool", "group", "email"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testbool", "name", "smtp.enabled"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testbool", "value", "false"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testbool", "type", "BOOLEAN"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testbool", "description", "Flag to enable/disable SMTP"),
					//
					resource.TestCheckResourceAttr("dependencytrack_config_property.teststring", "id", "email/subject.prefix"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.teststring", "group", "email"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.teststring", "name", "subject.prefix"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.teststring", "value", "TF Test with Update"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.teststring", "type", "STRING"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.teststring", "description", "The Prefix Subject email to use"),
					//
					resource.TestCheckResourceAttr("dependencytrack_config_property.testencrypted", "group", "email"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testencrypted", "name", "smtp.password"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testencrypted", "value", "TEST_PASSWORD"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testencrypted", "type", "ENCRYPTEDSTRING"),
					resource.TestCheckResourceAttr("dependencytrack_config_property.testencrypted", "description", "The optional password for the username used for authentication"),
				),
			},
		},
	})
}
