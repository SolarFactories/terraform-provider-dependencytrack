package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccRepositoryResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "dependencytrack_repository" "test_auth" {
	type = "GITHUB"
	identifier = "Test_Repository"
	url = "https://localhost"
	precedence = 2
	enabled = true
	internal = false
	username = "Test_Username"
	password = "Test_Password"
}`,
				/*resource "dependencytrack_repository" "test_unauth" {
				  	type = "GITHUB"
				  	identifier = "Test_Repository_Unauth"
				  	url = "https://localhost"
				  	precedence = 3
				  	enabled = true
				  	internal = false
				  	username = ""
				  	password = ""
				  }
				  `,*/

				Check: resource.ComposeAggregateTestCheckFunc(
					// Authenticated
					resource.TestCheckResourceAttrSet("dependencytrack_repository.test_auth", "id"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "type", "GITHUB"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "identifier", "Test_Repository"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "url", "https://localhost"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "precedence", "2"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "enabled", "true"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "internal", "false"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "username", "Test_Username"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "password", "Test_Password"),
					// Unauthenticated
					/*resource.TestCheckResourceAttrSet("dependencytrack_repository.test_unauth", "id"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_unauth", "type", "GITHUB"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_unauth", "identifier", "Test_Repository_Unauth"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_unauth", "url", "https://localhost"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_unauth", "precedence", "3"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_unauth", "enabled", "true"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_unauth", "internal", "false"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_unauth", "username", ""),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_unauth", "password", ""),*/
				),
			},
			// ImportState testing
			{
				ResourceName:      "dependencytrack_repository.test_auth",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "dependencytrack_repository" "test_auth" {
	type = "GITHUB"
	identifier = "Test_Repository_With_Change"
	url = "https://localhost"
	precedence = 2
	enabled = true
	internal = false
	username = "Test_Username"
	password = "Test_Password_With_Change"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_repository.test_auth", "id"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "type", "GITHUB"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "identifier", "Test_Repository_With_Change"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "url", "https://localhost"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "precedence", "2"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "enabled", "true"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "internal", "false"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "username", "Test_Username"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test_auth", "password", "Test_Password_With_Change"),
				),
			},
		},
	})
}
