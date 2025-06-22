package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTagResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_tag" "test" {
	name = "TagTagsTag"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag.test", "id", "TestTagsTag"),
					resource.TestCheckResourceAttr("dependencytrack_tag.test", "name", "TestTagsTag"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_tag.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `

resource "dependencytrack_tag" "test" {
	name = "TagTagsTagWithChange"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag.test", "id", "TestTagsTagWithChange"),
					resource.TestCheckResourceAttr("dependencytrack_tag.test", "name", "TestTagsTagWithChange"),
				),
			},
		},
	})
}
