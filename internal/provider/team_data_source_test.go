package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccTeamDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "dependencytrack_team" "test" {
	name = "Administrators"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dependencytrack_team.test", "id"),
					resource.TestCheckResourceAttr("data.dependencytrack_team.test", "name", "Administrators"),
					//
					resource.TestCheckResourceAttr("data.dependencytrack_team.test", "permissions.#", "14"),
					//
					resource.TestCheckResourceAttr("data.dependencytrack_team.test", "permissions.0.name", "ACCESS_MANAGEMENT"),
					resource.TestCheckResourceAttr("data.dependencytrack_team.test", "permissions.0.description", "Allows the management of users, teams, and API keys"),
				),
			},
		},
	})
}
