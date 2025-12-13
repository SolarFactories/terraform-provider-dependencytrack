package provider

import (
	"context"
	"fmt"
	"slices"
	"strings"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

type (
	teamPermissionsResource struct {
		client *dtrack.Client
		semver *Semver
	}

	teamPermissionsResourceModel struct {
		TeamID      types.String   `tfsdk:"team"`
		Permissions []types.String `tfsdk:"permissions"`
	}
)

func NewTeamPermissionsResource() resource.Resource {
	return &teamPermissionsResource{}
}

func (*teamPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team_permissions"
}

func (*teamPermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the attachment of Permissions to a Team. Conflicts with `dependencytrack_team_permission`.",
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
	teamID, errDiag := TryParseUUID(plan.TeamID, LifecycleCreate, path.Root("team"))
	if errDiag != nil {
		resp.Diagnostics.Append(errDiag)
		return
	}

	teamInfo, err := r.client.Team.Get(ctx, teamID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Create, unable to request current team info for: "+teamID.String(),
			"Error from: "+err.Error(),
		)
		return
	}

	desiredPermissions := Map(plan.Permissions, types.String.ValueString)
	currentPermissions := Map(teamInfo.Permissions, func(current dtrack.Permission) string { return current.Name })

	desiredPermissionMap := make(map[string]bool)
	currentPermissionMap := make(map[string]bool)

	for _, permission := range currentPermissions {
		currentPermissionMap[permission] = true
	}
	for _, permission := range desiredPermissions {
		desiredPermissionMap[permission] = true
	}

	addPermissions := []string{}
	removePermissions := []string{}
	for permission := range desiredPermissionMap {
		if !currentPermissionMap[permission] {
			addPermissions = append(addPermissions, permission)
		}
	}
	for permission := range currentPermissionMap {
		if !desiredPermissionMap[permission] {
			removePermissions = append(removePermissions, permission)
		}
	}

	tflog.Debug(ctx, "Creating Team Permissions", map[string]any{
		"team":    teamID.String(),
		"current": currentPermissions,
		"desired": desiredPermissions,
	})

	finalPermissions := r.updatePermissions(ctx, &resp.Diagnostics, teamInfo, addPermissions, removePermissions)
	if resp.Diagnostics.HasError() {
		return
	}
	finalPermissionsStrings := Map(finalPermissions, func(permission dtrack.Permission) string {
		return permission.Name
	})

	sortedDesiredPermissions := slices.SortedStableFunc(slices.Values(desiredPermissions), strings.Compare)
	sortedFinalStringsPermissions := slices.SortedStableFunc(slices.Values(finalPermissionsStrings), strings.Compare)

	var statePermissions = finalPermissionsStrings
	if slices.Equal(sortedDesiredPermissions, sortedFinalStringsPermissions) {
		statePermissions = desiredPermissions
	}

	plan = teamPermissionsResourceModel{
		TeamID:      types.StringValue(teamID.String()),
		Permissions: Map(statePermissions, types.StringValue),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created Team Permissions", map[string]any{
		"team":        plan.TeamID.ValueString(),
		"permissions": statePermissions,
	})
}

func (r *teamPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state.
	var state teamPermissionsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refresh.
	teamID, errDiag := TryParseUUID(state.TeamID, LifecycleRead, path.Root("team"))
	if errDiag != nil {
		resp.Diagnostics.Append(errDiag)
		return
	}
	tflog.Debug(ctx, "Reading Team Permissions", map[string]any{
		"team":          teamID.String(),
		"permissions.#": len(state.Permissions),
	})
	team, err := r.client.Team.Get(ctx, teamID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get updated team",
			"Error with reading team: "+teamID.String()+", in original error: "+err.Error(),
		)
		return
	}

	storedPermissionStrings := Map(state.Permissions, types.String.ValueString)
	updatedPermissionStrings := Map(team.Permissions, func(permission dtrack.Permission) string {
		return permission.Name
	})

	sortedUpdatedPermissions := slices.SortedStableFunc(slices.Values(updatedPermissionStrings), strings.Compare)
	sortedStoredPermissions := slices.SortedStableFunc(slices.Values(storedPermissionStrings), strings.Compare)

	statePermissions := updatedPermissionStrings
	if slices.Equal(sortedUpdatedPermissions, sortedStoredPermissions) {
		statePermissions = storedPermissionStrings
	}

	state = teamPermissionsResourceModel{
		TeamID:      types.StringValue(teamID.String()),
		Permissions: Map(statePermissions, types.StringValue),
	}

	// Update state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Team Permissions", map[string]any{
		"team":        state.TeamID.ValueString(),
		"permissions": team.Permissions,
	})
}

func (r *teamPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get State.
	var plan teamPermissionsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, errDiag := TryParseUUID(plan.TeamID, LifecycleUpdate, path.Root("team"))
	if errDiag != nil {
		resp.Diagnostics.Append(errDiag)
		return
	}
	tflog.Debug(ctx, "Updating Team Permissions", map[string]any{
		"team":          team.String(),
		"permissions.#": len(plan.Permissions),
	})

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

	tflog.Debug(ctx, "Updating Team Permissions", map[string]any{
		"current": currentPermissions,
		"desired": desiredPermissions,
	})

	addPermissions, removePermissions := ListDeltas(currentPermissions, desiredPermissions)
	finalPermissions := r.updatePermissions(ctx, &resp.Diagnostics, teamInfo, addPermissions, removePermissions)
	if resp.Diagnostics.HasError() {
		return
	}

	plan = teamPermissionsResourceModel{
		TeamID: types.StringValue(team.String()),
		Permissions: Map(finalPermissions, func(permission dtrack.Permission) types.String {
			return types.StringValue(permission.Name)
		}),
	}

	// Update State.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated Team Permissions", map[string]any{
		"team":        plan.TeamID.ValueString(),
		"permissions": finalPermissions,
	})
}

func (r *teamPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state.
	var state teamPermissionsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK.
	team, errDiag := TryParseUUID(state.TeamID, LifecycleDelete, path.Root("team"))
	if errDiag != nil {
		resp.Diagnostics.Append(errDiag)
		return
	}
	tflog.Debug(ctx, "Deleting Team Permissions", map[string]any{
		"team":          team.String(),
		"permissions.#": len(state.Permissions),
	})
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
				"Within Delete, error removing team permission: "+permission.Name+"for team: "+team.String(),
				"Error from: "+err.Error(),
			)
			continue
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleted Team Permissions", map[string]any{
		"team": state.TeamID.ValueString(),
		"permissions": Map(state.Permissions, func(permission types.String) string {
			return permission.ValueString()
		}),
	})
}

func (r *teamPermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *teamPermissionsResource) updatePermissions(
	ctx context.Context, diags *diag.Diagnostics,
	team dtrack.Team, addPermissions, removePermissions []string,
) []dtrack.Permission {
	// Ensure that only permissions assigned and understood by DT end up in state, rather than user input.
	finalPermissions := team.Permissions
	for _, permissionName := range addPermissions {
		permission := dtrack.Permission{Name: permissionName}
		updatedTeam, err := r.client.Permission.AddPermissionToTeam(ctx, permission, team.UUID)
		if err != nil {
			diags.AddError(
				"Error adding team permission: "+permissionName+" for team: "+team.UUID.String(),
				"Error from: "+err.Error(),
			)
			continue
		}
		finalPermissions = updatedTeam.Permissions
	}
	for _, permissionName := range removePermissions {
		permission := dtrack.Permission{Name: permissionName}
		updatedTeam, err := r.client.Permission.RemovePermissionFromTeam(ctx, permission, team.UUID)
		if err != nil {
			diags.AddError(
				"Error removing team permission: "+permissionName+" for team: "+team.UUID.String(),
				"Error from: "+err.Error(),
			)
			continue
		}
		finalPermissions = updatedTeam.Permissions
	}
	return finalPermissions
}
