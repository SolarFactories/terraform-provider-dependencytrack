package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccConfigPropertiesResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "dependencytrack_config_properties" "test" {
	properties = [
		{
			group = "email"
			name = "smtp.enabled"
			value = "true"
			type = "BOOLEAN"
		},
		{
			group = "email"
			name = "subject.prefix"
			value = "TF Test"
			type = "STRING"
		}
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.#", "2"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.0.group", "email"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.0.name", "smtp.enabled"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.0.value", "true"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.0.type", "BOOLEAN"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.0.description", "Flag to enable/disable SMTP"),
					//
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.1.group", "email"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.1.name", "subject.prefix"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.1.value", "TF Test"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.1.type", "STRING"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.1.description", "The Prefix Subject email to use"),
				),
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "dependencytrack_config_properties" "test" {
	properties = [
		{
			group = "email"
			name = "smtp.enabled"
			value = "false"
			type = "BOOLEAN"
		},
		{
			group = "email"
			name = "subject.prefix"
			value = "TF Test With Update"
			type = "STRING"
		}
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.#", "2"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.0.group", "email"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.0.name", "smtp.enabled"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.0.value", "false"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.0.type", "BOOLEAN"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.0.description", "Flag to enable/disable SMTP"),
					//
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.1.group", "email"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.1.name", "subject.prefix"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.1.value", "TF Test With Update"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.1.type", "STRING"),
					resource.TestCheckResourceAttr("dependencytrack_config_properties.test", "properties.1.description", "The Prefix Subject email to use"),
				),
			},
		},
	})
}
