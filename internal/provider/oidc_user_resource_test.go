package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOIDCUserResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_user" "test" {
	username = "Test_Username"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_oidc_user.test", "id", "Test_Username"),
					resource.TestCheckResourceAttr("dependencytrack_oidc_user.test", "username", "Test_Username"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_oidc_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_user" "test" {
	username = "Test_Username"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dependencytrack_oidc_user.test", "id", "Test_Username"),
					resource.TestCheckResourceAttr("dependencytrack_oidc_user.test", "username", "Test_Username"),
				),
			},
		},
	})
}
