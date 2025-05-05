package provider

import (
	"context"
	"fmt"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource              = &teamPermissionResource{}
	_ resource.ResourceWithConfigure = &teamPermissionResource{}
)

func NewTeamPermissionResource() resource.Resource {
	return &teamPermissionResource{}
}

type (
	teamPermissionResource struct {
		client *dtrack.Client
		semver *Semver
	}

	teamPermissionResourceModel struct {
		TeamID     types.String `tfsdk:"team"`
		Permission types.String `tfsdk:"permission"`
	}
)

func (*teamPermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_permission"
}

func (*teamPermissionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the attachment of a Permission to a Team. Conflicts with `dependencytrack_team_permissions`.",
		Attributes: map[string]schema.Attribute{
			"team": schema.StringAttribute{
				Description: "UUID for the Team for which to manage the permission.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permission": schema.StringAttribute{
				Description: "Permission name to attach / detach from the Team.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *teamPermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamPermissionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	team, diag := TryParseUUID(plan.TeamID, LifecycleCreate, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	permission := dtrack.Permission{
		Name: plan.Permission.ValueString(),
	}
	tflog.Debug(ctx, "Creating Team Permission", map[string]any{
		"team":       team.String(),
		"permission": permission.Name,
	})
	_, err := r.client.Permission.AddPermissionToTeam(ctx, permission, team)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating team permission",
			"Unexpected error: "+err.Error(),
		)
		return
	}
	plan.TeamID = types.StringValue(team.String())
	plan.Permission = types.StringValue(permission.Name)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created Team Permission", map[string]any{
		"team":       plan.TeamID.ValueString(),
		"permission": plan.Permission.ValueString(),
	})
}

func (r *teamPermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state.
	var state teamPermissionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refresh.
	teamID, diag := TryParseUUID(state.TeamID, LifecycleRead, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	tflog.Debug(ctx, "Read Team Permission", map[string]any{
		"team":       teamID.String(),
		"permission": state.Permission.ValueString(),
	})
	team, err := r.client.Team.Get(ctx, teamID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get updated team",
			"Error with reading team: "+teamID.String()+", in original error: "+err.Error(),
		)
		return
	}
	permission, err := Find(team.Permissions, func(permission dtrack.Permission) bool {
		return permission.Name == state.Permission.ValueString()
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to identify team permission",
			"Unexpected Error from: "+err.Error(),
		)
		return
	}
	state.TeamID = types.StringValue(team.UUID.String())
	state.Permission = types.StringValue(permission.Name)

	// Update state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Team Permission", map[string]any{
		"team":       state.TeamID.ValueString(),
		"permission": state.Permission.ValueString(),
	})
}

func (*teamPermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Nothing to Update. This resource only has Create, Delete actions.
	// Get State.
	var plan teamPermissionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update State.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated Team Permission", map[string]any{
		"team":       plan.TeamID.ValueString(),
		"permission": plan.Permission.ValueString(),
	})
}

func (r *teamPermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state.
	var state teamPermissionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK.
	team, diag := TryParseUUID(state.TeamID, LifecycleDelete, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	permission := dtrack.Permission{
		Name: state.Permission.ValueString(),
	}

	// Execute.
	tflog.Debug(ctx, "Deleting Team Permission", map[string]any{
		"team":       team.String(),
		"permission": permission.Name,
	})
	_, err := r.client.Permission.RemovePermissionFromTeam(ctx, permission, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete team permission",
			"Unexpected error when trying to delete team permission: "+team.String()+", error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted Team Permission", map[string]any{
		"team":       state.TeamID.ValueString(),
		"permission": state.Permission.ValueString(),
	})
}

func (r *teamPermissionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = clientInfoData.client
	r.semver = clientInfoData.semver
}
