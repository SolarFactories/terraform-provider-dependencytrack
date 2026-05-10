package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPermissionsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "dependencytrack_permissions" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dependencytrack_permissions.test", "permissions.#"),
					resource.TestCheckResourceAttr("data.dependencytrack_permissions.test", "permissions.0.name", "ACCESS_MANAGEMENT"),
					resource.TestCheckResourceAttr(
						"data.dependencytrack_permissions.test",
						"permissions.0.description",
						"Allows the management of users, teams, and API keys",
					),
				),
			},
		},
	})
}
