package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOidcGroupMappingResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_oidc_group" "test" {
	name = "Test_Group"
}
resource "dependencytrack_oidc_group_mapping" "test" {
	team = dependencytrack_team.test.id
	group = dependencytrack_oidc_group.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_oidc_group_mapping.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_oidc_group_mapping.test", "team",
						"dependencytrack_team.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_oidc_group_mapping.test", "group",
						"dependencytrack_oidc_group.test", "id",
					),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_oidc_group_mapping.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_oidc_group" "test" {
	name = "Test_Group"
}
resource "dependencytrack_oidc_group" "test2" {
	name = "Test_Group_2"
}
resource "dependencytrack_oidc_group_mapping" "test" {
	team = dependencytrack_team.test.id
	group = dependencytrack_oidc_group.test2.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_oidc_group_mapping.test", "id"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_oidc_group_mapping.test", "team",
						"dependencytrack_team.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_oidc_group_mapping.test", "group",
						"dependencytrack_oidc_group.test2", "id",
					),
				),
			},
		},
	})
}
