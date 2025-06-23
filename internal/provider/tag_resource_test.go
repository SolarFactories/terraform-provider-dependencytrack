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
	name = "test_tags_tag"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag.test", "id", "test_tags_tag"),
					resource.TestCheckResourceAttr("dependencytrack_tag.test", "name", "test_tags_tag"),
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
	name = "test_tags_tag_with_change"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_tag.test", "id", "test_tags_tag_with_change"),
					resource.TestCheckResourceAttr("dependencytrack_tag.test", "name", "test_tags_tag_with_change"),
				),
			},
		},
	})
}
