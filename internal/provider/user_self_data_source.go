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
	_ datasource.DataSource              = &userSelfDataSource{}
	_ datasource.DataSourceWithConfigure = &userSelfDataSource{}
)

type (
	userSelfDataSource struct {
		client *dtrack.Client
		semver *Semver
	}

	userSelfDataSourceModel struct {
		Username    types.String   `tfsdk:"username"`
		Email       types.String   `tfsdk:"email"`
		Name        types.String   `tfsdk:"name"`
		Teams       []types.String `tfsdk:"teams"`
		Permissions []types.String `tfsdk:"permissions"`
		ID          types.Int64    `tfsdk:"id"`
	}
)

func NewUserSelfDataSource() datasource.DataSource {
	return &userSelfDataSource{}
}

func (*userSelfDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_self"
}

func (*userSelfDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Obtain current authenticated user.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "Username of the current User.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email address of the current User.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the current User.",
				Computed:    true,
			},
			"teams": schema.ListAttribute{
				Description: "UUID's of teams that the current User is a member of.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"permissions": schema.ListAttribute{
				Description: "Permissions assigned directly to the current User.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"id": schema.Int64Attribute{
				Description: "ID of the current User.",
				Computed:    true,
			},
		},
	}
}

func (d *userSelfDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state userSelfDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Retrieving current user")

	user, err := d.client.User.GetSelf(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to retrieve current user",
			"Unexpected error within: "+err.Error(),
		)
		return
	}
	fmt.Printf("ID: %v\n", user.Id)
	newState := userSelfDataSourceModel{
		ID:          types.Int64Value(user.Id),
		Username:    types.StringValue(user.Username),
		Email:       types.StringValue(user.Email),
		Name:        types.StringValue(user.Name),
		Teams:       Map(user.Teams, func(team dtrack.Team) types.String { return types.StringValue(team.UUID.String()) }),
		Permissions: Map(user.Permissions, func(permission dtrack.Permission) types.String { return types.StringValue(permission.Name) }),
	}
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Retrieved current user")
}

func (d *userSelfDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
