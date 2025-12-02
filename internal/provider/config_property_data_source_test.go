package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConfigPropertyDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "dependencytrack_config_property" "test" {
	group = "general"
	name = "base.url"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dependencytrack_config_property.test", "group", "general"),
					resource.TestCheckResourceAttr("data.dependencytrack_config_property.test", "name", "base.url"),
					resource.TestCheckResourceAttr("data.dependencytrack_config_property.test", "value", ""),
					resource.TestCheckResourceAttr("data.dependencytrack_config_property.test", "type", "URL"),
					resource.TestCheckResourceAttr("data.dependencytrack_config_property.test", "description",
						"URL used to construct links back to Dependency-Track from external systems",
					),
				),
			},
		},
	})
}
