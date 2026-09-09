package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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

func TestAccPolicyResourceRegression236(t *testing.T) {
	// Regression test for https://github.com/SolarFactories/terraform-provider-dependencytrack/issues/236
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create initial Policy.
			{
				Config: providerConfig + `
resource "dependencytrack_policy" "test" {
	name = "Test_Policy_236"
	operator = "ANY"
	violation = "FAIL"
}
`,
			},
			// Duplicate reference to Policy.
			{
				Config: providerConfig + `
resource "dependencytrack_policy" "test" {
	name = "Test_Policy_236"
	operator = "ANY"
	violation = "FAIL"
}
resource "dependencytrack_policy" "test2" {
	name = "Test_Policy_236"
	operator = "ANY"
	violation = "FAIL"
}
import {
	to = dependencytrack_policy.test2
	id = dependencytrack_policy.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"dependencytrack_policy.test", "id",
						"dependencytrack_policy.test2", "id",
					),
				),
			},
			// Delete one.
			{
				Config: providerConfig + `
resource "dependencytrack_policy" "test2" {
	name = "Test_Policy_236"
	operator = "ANY"
	violation = "FAIL"
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("dependencytrack_policy.test2", "Create"),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
