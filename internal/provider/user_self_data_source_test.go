package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
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
		callback      func(string)
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
	fmt.Printf("Extracting token: [%s]\n", token)
	if extractor.token != nil {
		*extractor.token = token
	}
	if extractor.callback != nil {
		extractor.callback(token)
	}
}

func TestAccUserSelfDataSource(t *testing.T) {
	var token string
	tokenChecker := tokenExtractor{
		token: &token,
		callback: func(token string) {
			err := os.Setenv("TF_VAR_provider_bearer_token", token)
			if err != nil {
				panic(err)
			}
		},
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

data "dependencytrack_user_login" "test" {
	username = dependencytrack_user.user.username
	password = dependencytrack_user.user.password
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

variable "provider_bearer_token" {
	type = string
	sensitive = true
}

provider "dependencytrack" {
	host = local.provider_host
	root_ca = local.provider_root_ca
	mtls = local.provider_mtls

	auth = {
		type = "BEARER"
		bearer = var.provider_bearer_token
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
