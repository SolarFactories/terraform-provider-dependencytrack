package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserLoginDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "dependencytrack_user_login" "test" {
	username = "admin"
	password = "pipeline"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dependencytrack_user_login.test", "token"),
					resource.TestCheckResourceAttr("data.dependencytrack_user_login.test", "username", "admin"),
					resource.TestCheckResourceAttr("data.dependencytrack_user_login.test", "password", ""),
				),
			},
		},
	})
}
