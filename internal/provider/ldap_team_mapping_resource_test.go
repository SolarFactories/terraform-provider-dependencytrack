package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLDAPMappingResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_ldap_team_mapping" "test" {
	team = dependencytrack_team.test.id
	distinguished_name = "test.mapping.ldap.dependencytrack.solarfactories"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_ldap_team_mapping.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_ldap_team_mapping.test", "distinguished_name", "test.mapping.ldap.dependencytrack.solarfactories"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_ldap_team_mapping.test", "team",
						"dependencytrack_team.test", "id",
					),
				),
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_ldap_team_mapping" "test" {
	team = dependencytrack_team.test.id
	distinguished_name = "test2.mapping.ldap.dependencytrack.solarfactories"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_ldap_team_mapping.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_ldap_team_mapping.test", "distinguished_name", "test2.mapping.ldap.dependencytrack.solarfactories"),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_ldap_team_mapping.test", "team",
						"dependencytrack_team.test", "id",
					),
				),
			},
		},
	})
}
