package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOidcLoginDataSource(t *testing.T) {
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
data "dependencytrack_oidc_login" "test" {
	id_token = var.oidc_id_token
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dependencytrack_oidc_login.test", "token"),
				),
			},
		},
	})
}
