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
	_ datasource.DataSource              = &userLoginDataSource{}
	_ datasource.DataSourceWithConfigure = &userLoginDataSource{}
)

type (
	userLoginDataSource struct {
		client *dtrack.Client
		semver *Semver
	}

	userLoginDataSourceModel struct {
		Token    types.String `tfsdk:"token"`
		Username types.String `tfsdk:"username"`
		Password types.String `tfsdk:"password"`
	}
)

func NewUserLoginDataSource() datasource.DataSource {
	return &userLoginDataSource{}
}

func (*userLoginDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_login"
}

func (*userLoginDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Authenticate using Username and password.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Description: "DependencyTrack bearer token.",
				Sensitive:   true,
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "Username within DependencyTrack.",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for Managed User within DependencyTrack.",
				Sensitive:   true,
				Required:    true,
			},
		},
	}
}

func (d *userLoginDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state userLoginDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Authenticating using Username and Password")

	username := state.Username.ValueString()
	password := state.Password.ValueString()
	token, err := d.client.User.Login(ctx, username, password)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to Login with Username and Password",
			"Unexpected error within: "+err.Error(),
		)
		return
	}

	newState := userLoginDataSourceModel{
		Token:    types.StringValue(token),
		Username: state.Username,
		Password: types.StringNull(),
	}

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Authenticated using Username and Password")
}

func (d *userLoginDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
