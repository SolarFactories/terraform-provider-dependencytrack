package provider

import (
	"context"
	"errors"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

var _ statecheck.StateCheck = tokenExtractor{}

type (
	tokenExtractor struct {
		token         *string
		resAddress    string
		attributePath tfjsonpath.Path
	}
)

func (extractor tokenExtractor) CheckState(_ context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	r, err := Find(req.State.Values.RootModule.Resources, func(r *tfjson.StateResource) bool {
		return extractor.resAddress == r.Address
	})
	if err != nil {
		resp.Error = err
		return
	}
	result, err := tfjsonpath.Traverse((*r).AttributeValues, extractor.attributePath)
	if err != nil {
		resp.Error = err
		return
	}
	token, ok := result.(string)
	if !ok {
		resp.Error = errors.New("expected string result, but received another type")
		return
	}
	*extractor.token = token
}

func TestAccUserSelfDataSource(t *testing.T) {
	var token string
	tokenChecker := tokenExtractor{
		token:         &token,
		resAddress:    "data.dependencytrack_user_login.test",
		attributePath: tfjsonpath.New("token"),
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + `
resource "dependencytrack_user" "user" {
	username = "Test_User_Self"
	fullname = "Test User"
	email = "test@example.com"
	force_password_change = false
	password = "PASSWORD"
}

provider "dependencytrack" {
	host = local.provider_host
	root_ca = local.provider_root_ca
	mtls = local.provider_mtls

	auth = {
		type = "NONE"
	}
	alias = "unauthenticated"
}

data "dependencytrack_user_login" "test" {
	username = dependencytrack_user.user.username
	password = dependencytrack_user.user.password
	provider = dependencytrack.unauthenticated
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					tokenChecker,
				},
			},
			{
				Config: providerConfig + `
resource "dependencytrack_user" "user" {
	username = "Test_User_Self"
	fullname = "Test User"
	email = "test@example.com"
	force_password_change = false
	password = "PASSWORD"
}

provider "dependencytrack" {
	host = local.provider_host
	root_ca = local.provider_root_ca
	mtls = local.provider_mtls

	auth = {
		type = "BEARER"
		bearer = "` + token + `"
	}
	alias = "managed"
}

data "dependencytrack_user_self" "test" {
	provider = dependencytrack.managed
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}
