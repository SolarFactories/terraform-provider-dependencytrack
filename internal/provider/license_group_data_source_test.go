package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestLicenseGroupDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
data "dependencytrack_license_group" "test" {
	name = "Permissive"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.dependencytrack_license_group.test", "id"),
					resource.TestCheckResourceAttr("data.dependencytrack_license_group.test", "name", "Permissive"),
					resource.TestCheckResourceAttrSet("data.dependencytrack_license_group.test", "risk_weight"),
					resource.TestCheckResourceAttrSet("data.dependencytrack_license_group.test", "licenses.#"),
					resource.TestCheckResourceAttrSet("data.dependencytrack_license_group.test", "licenses.0.uuid"),
					resource.TestCheckResourceAttrSet("data.dependencytrack_license_group.test", "licenses.0.name"),
					resource.TestCheckResourceAttrSet("data.dependencytrack_license_group.test", "licenses.0.spdx_id"),
				),
			},
		},
	})
}
