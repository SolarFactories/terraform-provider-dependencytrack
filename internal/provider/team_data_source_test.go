package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
					// TODO: Introduce new data source for dependencytrack_about, or similar, which will expose API version,
					// and static information, such as defined permissions within API,
					// avoiding issues of different versions of API introducing new permissions,
					// and so throwing off value here. 14 in 4.12.7, 12 in 4.11.7
					//resource.TestCheckResourceAttr("data.dependencytrack_team.test", "permissions.#", "14"),.
					//
					resource.TestCheckResourceAttr("data.dependencytrack_team.test", "permissions.0.name", "ACCESS_MANAGEMENT"),
					resource.TestCheckResourceAttr("data.dependencytrack_team.test", "permissions.0.description", "Allows the management of users, teams, and API keys"),
				),
			},
		},
	})
}
