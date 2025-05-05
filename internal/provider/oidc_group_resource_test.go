package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccOidcGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_group" "test" {
	name = "Test_Group"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_oidc_group.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_oidc_group.test", "name", "Test_Group"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_oidc_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_group" "test" {
	name = "Test_Group_2"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_oidc_group.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_oidc_group.test", "name", "Test_Group_2"),
				),
			},
		},
	})
}
