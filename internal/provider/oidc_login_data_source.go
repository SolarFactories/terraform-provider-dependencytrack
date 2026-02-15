package provider

import (
	"context"
	"fmt"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Interface impl check.
var (
	_ datasource.DataSource              = &oidcLoginDataSource{}
	_ datasource.DataSourceWithConfigure = &oidcLoginDataSource{}
)

type (
	oidcLoginDataSource struct {
		client *dtrack.Client
		semver *Semver
	}

	oidcLoginDataSourceModel struct {
		Token       types.String `tfsdk:"token"`
		IDToken     types.String `tfsdk:"id_token"`
		AccessToken types.String `tfsdk:"access_token"`
	}
)

func NewOidcLoginDataSource() datasource.DataSource {
	return &oidcLoginDataSource{}
}

func (*oidcLoginDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oidc_login"
}

func (*oidcLoginDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Authenticate using OIDC Tokens.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Description: "DependencyTrack bearer token.",
				Sensitive:   true,
				Computed:    true,
			},
			"id_token": schema.StringAttribute{
				Description: "OIDC ID Token from Identity Provider.",
				Sensitive:   true,
				Required:    true,
			},
			"access_token": schema.StringAttribute{
				Description: "OIDC Access Token from Identity Provider. Optional if all required fields are present in ID Token.",
				Sensitive:   true,
				Optional:    true,
			},
		},
	}
}

func (d *oidcLoginDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state oidcLoginDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Authenticating using OIDC Tokens")

	token, err := d.client.OIDC.Login(ctx, dtrack.OIDCTokens{
		ID:     state.IDToken.ValueString(),
		Access: state.AccessToken.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to fetch OIDC availability.",
			"Unexpected error within: "+err.Error(),
		)
		return
	}
	newState := oidcLoginDataSourceModel{
		Token:       types.StringValue(token),
		IDToken:     state.IDToken,
		AccessToken: state.AccessToken,
	}
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Authenticated using OIDC Tokens")
}

func (d *oidcLoginDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clientInfoData, ok := req.ProviderData.(clientInfo)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Configure Type",
			fmt.Sprintf("Expected provider.clientInfo, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = clientInfoData.client
	d.semver = clientInfoData.semver
}
