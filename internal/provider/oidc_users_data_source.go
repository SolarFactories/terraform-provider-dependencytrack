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
	_ datasource.DataSource              = &oidcUsersDataSource{}
	_ datasource.DataSourceWithConfigure = &oidcUsersDataSource{}
)

type (
	oidcUsersDataSource struct {
		client *dtrack.Client
		semver *Semver
	}

	oidcUsersDataSourceModel struct {
		Users      []oidcUsersUserModel `tfsdk:"users"`
		TotalCount types.Int32          `tfsdk:"total_count"`
	}

	oidcUsersUserModel struct {
		Username    types.String             `tfsdk:"username"`
		Teams       []oidcUsersUserTeamModel `tfsdk:"teams"`
		Permissions []types.String           `tfsdk:"permissions"`
	}

	oidcUsersUserTeamModel struct {
		ID   types.String `tfsdk:"id"`
		Name types.String `tfsdk:"name"`
	}
)

func NewOidcUsersDataSource() datasource.DataSource {
	return &oidcUsersDataSource{}
}

func (*oidcUsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oidc_users"
}

func (*oidcUsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetch all OIDC Users.",
		Attributes: map[string]schema.Attribute{
			"total_count": schema.Int32Attribute{
				Description: "Total Count of OIDC Users.",
				Computed:    true,
			},
			"users": schema.ListNestedAttribute{
				Description: "List of OIDC Users.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"teams": schema.ListNestedAttribute{
							Description: "List of teams of which the User is a member.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Description: "UUID of the team including the User.",
										Computed:    true,
									},
									"name": schema.StringAttribute{
										Description: "Name of the team including the User.",
										Computed:    true,
									},
								},
							},
						},
						"permissions": schema.ListAttribute{
							Description: "Name of permissions assigned to the User.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *oidcUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state oidcUsersDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading OIDC Users")

	users, err := d.client.OIDC.GetAllUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to fetch OIDC Users.",
			"Unexpected error within: "+err.Error(),
		)
		return
	}

	newState := oidcUsersDataSourceModel{
		TotalCount: types.Int32Value(int32(users.TotalCount)),
		Users: Map(users.Items, func(user dtrack.OIDCUser) oidcUsersUserModel {
			return oidcUsersUserModel{
				Username: types.StringValue(user.Username),
				Teams: Map(user.Teams, func(team dtrack.Team) oidcUsersUserTeamModel {
					return oidcUsersUserTeamModel{
						Name: types.StringValue(team.Name),
						ID:   types.StringValue(team.UUID.String()),
					}
				}),
				Permissions: Map(user.Permissions, func(perm dtrack.Permission) types.String {
					return types.StringValue(perm.Name)
				}),
			}
		}),
	}

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read OIDC Users", map[string]any{
		"total_count": newState.TotalCount.ValueInt32(),
		"users.#":     len(newState.Users),
	})
}

func (d *oidcUsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
