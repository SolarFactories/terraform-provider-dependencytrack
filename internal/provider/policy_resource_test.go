package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccPolicyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_policy" "test" {
	name = "Test_Policy"
	operator = "ANY"
	violation = "FAIL"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_policy.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_policy.test", "name", "Test_Policy"),
					resource.TestCheckResourceAttr("dependencytrack_policy.test", "operator", "ANY"),
					resource.TestCheckResourceAttr("dependencytrack_policy.test", "violation", "FAIL"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "dependencytrack_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing.
			{
				Config: providerConfig + `
resource "dependencytrack_policy" "test" {
	name = "Test_Policy_2"
	operator = "ANY"
	violation = "FAIL"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("dependencytrack_policy.test", "id"),
					resource.TestCheckResourceAttr("dependencytrack_policy.test", "name", "Test_Policy_2"),
					resource.TestCheckResourceAttr("dependencytrack_policy.test", "operator", "ANY"),
					resource.TestCheckResourceAttr("dependencytrack_policy.test", "violation", "FAIL"),
				),
			},
		},
	})
}
