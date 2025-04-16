package provider

import (
	"context"
	"fmt"
	"github.com/google/uuid"

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
	_ resource.Resource              = &teamPermissionsResource{}
	_ resource.ResourceWithConfigure = &teamPermissionsResource{}
)

func NewTeamPermissionsResource() resource.Resource {
	return &teamPermissionsResource{}
}

type teamPermissionsResource struct {
	client *dtrack.Client
	semver *Semver
}

type teamPermissionsResourceModel struct {
	TeamID      types.String   `tfsdk:"team"`
	Permissions []types.String `tfsdk:"permissions"`
}

func (r *teamPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_permissions"
}

func (r *teamPermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the attachment of a Permission to a Team.",
		Attributes: map[string]schema.Attribute{
			"team": schema.StringAttribute{
				Description: "UUID for the Team for which to manage the permissions.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permissions": schema.ListAttribute{
				Description: "Alphabetically sorted Permissions for team. Conflicts with `dependencytrack_team_permission`. See DependencyTrack for allowed values.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *teamPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamPermissionsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	team, err := uuid.Parse(plan.TeamID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("team"),
			"Within Create, unable to parse team into UUID",
			"Error from: "+err.Error(),
		)
		return
	}

	teamInfo, err := r.client.Team.Get(ctx, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Create, unable to request current team info for: "+team.String(),
			"Error from: "+err.Error(),
		)
		return
	}

	desiredPermissions := Map(plan.Permissions, func(desired types.String) string { return desired.ValueString() })
	currentPermissions := Map(teamInfo.Permissions, func(current dtrack.Permission) string { return current.Name })

	finalPermissions := teamInfo.Permissions // Ensure that only permissions assigned and understood by DT end up in state, rather than user input
	addPermissions, removePermissions := ListDeltas(currentPermissions, desiredPermissions)
	for _, permissionName := range addPermissions {
		permission := dtrack.Permission{
			Name: permissionName,
		}
		updatedTeam, err := r.client.Permission.AddPermissionToTeam(ctx, permission, team)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error adding team permission: "+permission.Name+" for team: "+team.String(),
				"Error from: "+err.Error(),
			)
		}
		finalPermissions = updatedTeam.Permissions
	}
	for _, permissionName := range removePermissions {
		permission := dtrack.Permission{
			Name: permissionName,
		}
		updatedTeam, err := r.client.Permission.RemovePermissionFromTeam(ctx, permission, team)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error removing team permission: "+permission.Name+"for team: "+team.String(),
				"Error from: "+err.Error(),
			)
		}
		finalPermissions = updatedTeam.Permissions
	}
	if resp.Diagnostics.HasError() {
		return
	}

	plan = teamPermissionsResourceModel{
		TeamID: types.StringValue(team.String()),
		Permissions: Map(finalPermissions, func(permission dtrack.Permission) types.String {
			return types.StringValue(permission.Name)
		}),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Assigned permissions to team, with id: "+team.String())
}

func (r *teamPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state
	var state teamPermissionsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Refreshing team permissions for team: "+state.TeamID.ValueString())

	// Refresh
	id, err := uuid.Parse(state.TeamID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("team"),
			"Within Read, unable to parse team into UUID",
			"Error from: "+err.Error(),
		)
		return
	}
	team, err := r.client.Team.Get(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get updated team",
			"Error with reading team: "+id.String()+", in original error: "+err.Error(),
		)
		return
	}

	state = teamPermissionsResourceModel{
		TeamID: types.StringValue(id.String()),
		Permissions: Map(team.Permissions, func(permission dtrack.Permission) types.String {
			return types.StringValue(permission.Name)
		}),
	}

	// Update state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Refreshed permissions for team with id: "+state.TeamID.ValueString())
}

func (r *teamPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get State
	var plan teamPermissionsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := uuid.Parse(plan.TeamID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("team"),
			"Within Update, unable to parse team into UUID",
			"Error from: "+err.Error(),
		)
		return
	}

	teamInfo, err := r.client.Team.Get(ctx, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Update, unable to request current team info for: "+team.String(),
			"Error from: "+err.Error(),
		)
		return
	}

	desiredPermissions := Map(plan.Permissions, func(desired types.String) string { return desired.ValueString() })
	currentPermissions := Map(teamInfo.Permissions, func(current dtrack.Permission) string { return current.Name })

	finalPermissions := teamInfo.Permissions // Ensure that only permissions assigned and understood by DT end up in state, rather than user input
	addPermissions, removePermissions := ListDeltas(currentPermissions, desiredPermissions)
	for _, permissionName := range addPermissions {
		permission := dtrack.Permission{
			Name: permissionName,
		}
		updatedTeam, err := r.client.Permission.AddPermissionToTeam(ctx, permission, team)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error adding team permission: "+permission.Name+" for team: "+team.String(),
				"Error from: "+err.Error(),
			)
		}
		finalPermissions = updatedTeam.Permissions
	}
	for _, permissionName := range removePermissions {
		permission := dtrack.Permission{
			Name: permissionName,
		}
		updatedTeam, err := r.client.Permission.RemovePermissionFromTeam(ctx, permission, team)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error removing team permission: "+permission.Name+"for team: "+team.String(),
				"Error from: "+err.Error(),
			)
		}
		finalPermissions = updatedTeam.Permissions
	}
	if resp.Diagnostics.HasError() {
		return
	}

	plan = teamPermissionsResourceModel{
		TeamID: types.StringValue(team.String()),
		Permissions: Map(finalPermissions, func(permission dtrack.Permission) types.String {
			return types.StringValue(permission.Name)
		}),
	}

	// Update State
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated permissions for team with id: "+plan.TeamID.ValueString())
}

func (r *teamPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state
	var state teamPermissionsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK
	team, err := uuid.Parse(state.TeamID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("team"),
			"Within Delete, unable to parse UUID",
			"Error parsing UUID from: "+state.TeamID.ValueString()+", error: "+err.Error(),
		)
		return
	}
	teamInfo, err := r.client.Team.Get(ctx, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Delete, unable to request current team info for: "+team.String(),
			"Error from: "+err.Error(),
		)
		return
	}

	for _, permission := range teamInfo.Permissions {
		_, err = r.client.Permission.RemovePermissionFromTeam(ctx, permission, team)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error removing team permission: "+permission.Name+"for team: "+team.String(),
				"Error from: "+err.Error(),
			)
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleted permissions from team with id: "+team.String())
}

func (r *teamPermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clientInfo, ok := req.ProviderData.(clientInfo)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Configure Type",
			fmt.Sprintf("Expected provider.clientInfo, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = clientInfo.client
	r.semver = clientInfo.semver
}
