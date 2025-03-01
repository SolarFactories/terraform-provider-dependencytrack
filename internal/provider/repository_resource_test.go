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
resource "dependencytrack_repository" "test" {
	type = "GITHUB"
	identifier = "Test_Repository"
	url = "https://localhost"
	precedence = 2
	enabled = true
	internal = false
	username = "Test_Username"
	password = "Test_Password"
}
`,

				Check: resource.ComposeAggregateTestCheckFunc(
					// Authenticated
					resource.TestCheckResourceAttrSet("dependencytrack_repository.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "type", "GITHUB"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "identifier", "Test_Repository"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "url", "https://localhost"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "precedence", "2"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "internal", "false"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "username", "Test_Username"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "password", "Test_Password"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "dependencytrack_repository.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "dependencytrack_repository" "test" {
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
					resource.TestCheckResourceAttrSet("dependencytrack_repository.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "type", "GITHUB"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "identifier", "Test_Repository_With_Change"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "url", "https://localhost"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "precedence", "2"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "enabled", "true"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "internal", "false"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "username", "Test_Username"),
					resource.TestCheckResourceAttr("dependencytrack_repository.test", "password", "Test_Password_With_Change"),
				),
			},
		},
	})
}
