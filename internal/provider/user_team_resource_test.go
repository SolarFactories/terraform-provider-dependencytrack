package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserTeamResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team"
}
resource "dependencytrack_user" "test" {
	username = "Test_User"
	fullname = "Test User"
	email = "Test_User@example.com"
	password = "Test_User_Password"
}
resource "dependencytrack_user_team" "test" {
	username = dependencytrack_user.test.username
	team = dependencytrack_team.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_user_team.test", "username",
						"dependencytrack_user.test", "username",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_user_team.test", "team",
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
resource "dependencytrack_user" "test" {
	username = "Test_User_With_Change"
	fullname = "Test User"
	email = "Test_User@example.com"
	password = "Test_User_Password"
}
resource "dependencytrack_user_team" "test" {
	username = dependencytrack_user.test.username
	team = dependencytrack_team.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_user_team.test", "username",
						"dependencytrack_user.test", "username",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_user_team.test", "team",
						"dependencytrack_team.test", "id",
					),
				),
			},
		},
	})
}

func TestAccUserTeamResourceSSORegression177(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/177
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_team" "test" {
	name = "Test_Team_Regression177"
}

resource "dependencytrack_oidc_user" "test" {
	username = "Test_User_Regression177_OIDC"
}
resource "dependencytrack_ldap_user" "test" {
	username = "Test_User_Regression177_LDAP"
}

resource "dependencytrack_user_team" "oidc" {
	username = dependencytrack_oidc_user.test.username
	team = dependencytrack_team.test.id
}
resource "dependencytrack_user_team" "ldap" {
	username = dependencytrack_ldap_user.test.username
	team = dependencytrack_team.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_user_team.oidc", "username",
						"dependencytrack_oidc_user.test", "username",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_user_team.oidc", "team",
						"dependencytrack_team.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_user_team.ldap", "username",
						"dependencytrack_ldap_user.test", "username",
					),
					resource.TestCheckResourceAttrPair(
						"dependencytrack_user_team.ldap", "team",
						"dependencytrack_team.test", "id",
					),
				),
			},
		},
	})
}
