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
	_ datasource.DataSource              = &teamDataSource{}
	_ datasource.DataSourceWithConfigure = &teamDataSource{}
)

func NewTeamDataSource() datasource.DataSource {
	return &teamDataSource{}
}

type (
	teamDataSource struct {
		client *dtrack.Client
		semver *Semver
	}

	teamDataSourceModel struct {
		ID          types.String          `tfsdk:"id"`
		Name        types.String          `tfsdk:"name"`
		Permissions []teamPermissionModel `tfsdk:"permissions"`
	}

	teamPermissionModel struct {
		Name        types.String `tfsdk:"name"`
		Description types.String `tfsdk:"description"`
	}
)

func (d *teamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (d *teamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetch an existing Team by Name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the team located.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the team to find.",
				Required:    true,
			},
			"permissions": schema.ListNestedAttribute{
				Description: "Existing permissions within the Team.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Property Name for the Team Permission.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description for the Team Permission.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *teamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state teamDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Reading Team", map[string]any{
		"name": state.Name.ValueString(),
	})

	team, err := FindPaged(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.Team], error) {
			return d.client.Team.GetAll(ctx, po)
		},
		func(team dtrack.Team) bool {
			return team.Name == state.Name.ValueString()
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to fetch teams.",
			"Unexpected error from: "+err.Error(),
		)
		return
	}
	// Transform data into model.
	teamState := teamDataSourceModel{
		Name:        types.StringValue(team.Name),
		ID:          types.StringValue(team.UUID.String()),
		Permissions: make([]teamPermissionModel, 0),
	}
	for _, permission := range team.Permissions {
		teamState.Permissions = append(teamState.Permissions, teamPermissionModel{
			Name:        types.StringValue(permission.Name),
			Description: types.StringValue(permission.Description),
		})
		tflog.Debug(ctx, "Found Team Permission", map[string]any{
			"team":        team.UUID.String(),
			"name":        permission.Name,
			"description": permission.Description,
		})
	}
	// Update state.
	diags = resp.State.Set(ctx, &teamState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Team", map[string]any{
		"id":            teamState.ID.ValueString(),
		"name":          teamState.Name.ValueString(),
		"permissions.#": len(teamState.Permissions),
	})
}

func (d *teamDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
