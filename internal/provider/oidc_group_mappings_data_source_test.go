package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOidcGroupMappingsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_OIDC_Group_Mappings_DS_Team"
}
resource "dependencytrack_oidc_group" "test" {
	name = "Test_OIDC_Group_Mappings_DS_Group"
}
resource "dependencytrack_oidc_group_mapping" "test" {
	team = dependencytrack_team.test.id
	group = dependencytrack_oidc_group.test.id
}
data "dependencytrack_oidc_group_mappings" "test" {
	group = dependencytrack_oidc_group.test.id
	depends_on = [
		dependencytrack_oidc_group_mapping.test
	]
}

`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dependencytrack_oidc_group_mappings.test", "teams.#", "1"),
					resource.TestCheckResourceAttrPair(
						"data.dependencytrack_oidc_group_mappings.test", "teams.0.id",
						"dependencytrack_team.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.dependencytrack_oidc_group_mappings.test", "teams.0.name",
						"dependencytrack_team.test", "name",
					),
				),
			},
		},
	})
}

func TestAccOidcGroupMappingsDataSourceEmpty(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "dependencytrack_oidc_group" "test" {
	name = "Test_OIDC_Group_Mappings_DS_Empty_Group"
}
data "dependencytrack_oidc_group_mappings" "test" {
	group = dependencytrack_oidc_group.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dependencytrack_oidc_group_mappings.test", "teams.#", "0"),
				),
			},
		},
	})
}

func TestAccOidcGroupMappingsDataSourceOrder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "dependencytrack_team" "team_a" {
	name = "Test_OIDC_Group_Mappings_DS_Team_A"
}
resource "dependencytrack_team" "team_z" {
	name = "Test_OIDC_Group_Mappings_DS_Team_Z"
}
resource "dependencytrack_oidc_group" "test" {
	name = "Test_OIDC_Group_Mappings_DS_Order_Group"
}
resource "dependencytrack_oidc_group_mapping" "test_a" {
	team = dependencytrack_team.team_a.id
	group = dependencytrack_oidc_group.test.id
}
resource "dependencytrack_oidc_group_mapping" "test_z" {
	team = dependencytrack_team.team_z.id
	group = dependencytrack_oidc_group.test.id
}
data "dependencytrack_oidc_group_mappings" "test" {
	group = dependencytrack_oidc_group.test.id
	depends_on = [
		dependencytrack_oidc_group_mapping.test_a,
		dependencytrack_oidc_group_mapping.test_z,
	]
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.dependencytrack_oidc_group_mappings.test", "teams.#", "2"),
					resource.TestCheckResourceAttrPair(
						"data.dependencytrack_oidc_group_mappings.test", "teams.0.id",
						"dependencytrack_team.team_a", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.dependencytrack_oidc_group_mappings.test", "teams.0.name",
						"dependencytrack_team.team_a", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.dependencytrack_oidc_group_mappings.test", "teams.1.id",
						"dependencytrack_team.team_z", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.dependencytrack_oidc_group_mappings.test", "teams.1.name",
						"dependencytrack_team.team_z", "name",
					),
				),
			},
		},
	})
}
