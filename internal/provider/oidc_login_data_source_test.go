package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOidcLoginDataSource(t *testing.T) {
	// TODO: In GitHub actions, this is able to use a generated ID Token.
	//	This does not work well locally, so will require likely adding an IdP to Docker Compose Config, e.g. a minimal KeyCloak installation.
	//	For which, could then automatically obtain ID Token, using hashicorp/http provider.
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
variable "oidc_id_token" {
	type = string
	sensitive = true
	nullable = false
}
data "dependencytrack_oidc_login" {
	id_token = var.oidc_id_token
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dependencytrack_oidc_login", "token"),
				),
			},
		},
	})
}
