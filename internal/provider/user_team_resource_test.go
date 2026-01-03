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
resource "dependencytrack_managed_user" "test" {
	username = "Test_User"
	fullname = "Test User"
	email = "Test_User@example.com"
	password = "Test_User_Password"
}
resource "dependencytrack_user_team" "test" {
	username = dependencytrack_managed_user.test.username
	team = dependencytrack_team.test.id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_user_team.test", "username",
						"dependencytrack_managed_user.test", "username",
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
resource "dependencytrack_managed_user" "test" {
	username = "Test_User_With_Change"
	fullname = "Test User"
	email = "Test_User@example.com"
	password = "Test_User_Password"
}
resource "dependencytrack_user_team" "test" {
	username = dependencytrack_managed_user.test.username
	team = dependencytrack_team.test.id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_user_team.test", "username",
						"dependencytrack_managed_user.test", "username",
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
