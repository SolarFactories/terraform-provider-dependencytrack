package provider

import (
	"context"
	"fmt"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Interface impl check.
var (
	_ datasource.DataSource              = &oidcGroupMappingsDataSource{}
	_ datasource.DataSourceWithConfigure = &oidcGroupMappingsDataSource{}
)

type (
	oidcGroupMappingsDataSource struct {
		client *dtrack.Client
		semver *Semver
	}

	oidcGroupMappingsDataSourceModel struct {
		GroupID types.String                  `tfsdk:"group"`
		Teams   []oidcGroupMappingsTeamsModel `tfsdk:"teams"`
	}

	oidcGroupMappingsTeamsModel struct {
		ID   types.String `tfsdk:"id"`
		Name types.String `tfsdk:"name"`
	}
)

func NewOidcGroupMappingsDataSource() datasource.DataSource {
	return &oidcGroupMappingsDataSource{}
}

func (*oidcGroupMappingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oidc_group_mappings"
}

func (*oidcGroupMappingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetch all teams for an OIDC Group Mapping.",
		Attributes: map[string]schema.Attribute{
			"group": schema.StringAttribute{
				Description: "UUID for the OIDC Group.",
				Required:    true,
			},
			"teams": schema.ListNestedAttribute{
				Description: "List of teams mapped to the OIDC Group.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "UUID of the team mapped to the group.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the team mapped to the group.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *oidcGroupMappingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state oidcGroupMappingsDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID, diag := TryParseUUID(state.GroupID, LifecycleRead, path.Root("group"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	tflog.Debug(ctx, "Reading OIDC Group Mapping Teams", map[string]any{
		"group": groupID,
	})
	teams, err := d.client.OIDC.GetAllTeamsOf(ctx, dtrack.OIDCGroup{UUID: groupID, Name: ""})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to fetch OIDC Group Mapping Teams.",
			"Unexpected error within: "+err.Error(),
		)
		return
	}
	newState := oidcGroupMappingsDataSourceModel{
		GroupID: types.StringValue(groupID.String()),
		Teams: Map(teams, func(team dtrack.Team) oidcGroupMappingsTeamsModel {
			return oidcGroupMappingsTeamsModel{
				ID:   types.StringValue(team.UUID.String()),
				Name: types.StringValue(team.Name),
			}
		}),
	}
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read OIDC Group Mapping Teams", map[string]any{
		"group":   newState.GroupID.ValueString(),
		"teams.#": len(newState.Teams),
	})
}

func (d *oidcGroupMappingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
