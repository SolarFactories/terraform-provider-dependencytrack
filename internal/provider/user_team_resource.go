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
	_ resource.Resource              = &userTeamResource{}
	_ resource.ResourceWithConfigure = &userTeamResource{}
)

type (
	userTeamResource struct {
		client *dtrack.Client
		semver *Semver
	}

	userTeamResourceModel struct {
		Username types.String `tfsdk:"username"`
		TeamID   types.String `tfsdk:"team"`
	}
)

func NewUserTeamResource() resource.Resource {
	return &userTeamResource{}
}

func (*userTeamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_team"
}

func (*userTeamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a team membership for a user.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Description: "Username for which to manage Team membership.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team": schema.StringAttribute{
				Description: "UUID for the Team.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *userTeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan userTeamResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	username := plan.Username.ValueString()
	team, diag := TryParseUUID(plan.TeamID, LifecycleCreate, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	tflog.Debug(ctx, "Creating User Team Membership", map[string]any{
		"username": username,
		"team":     team.String(),
	})

	user, err := r.client.User.AddTeamToUser(ctx, username, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding user to team",
			"Error with user: "+username+", with team: "+team.String()+", in original error: "+err.Error(),
		)
		return
	}

	state := userTeamResourceModel{
		Username: types.StringValue(user.Username),
		TeamID:   types.StringValue(team.String()),
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created User Team Membership", map[string]any{
		"username": state.Username.ValueString(),
		"team":     state.TeamID.ValueString(),
	})
}

func (r *userTeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state userTeamResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := state.Username.ValueString()
	teamID, diag := TryParseUUID(state.TeamID, LifecycleRead, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	tflog.Debug(ctx, "Reading User Team Membership", map[string]any{
		"username": username,
		"team":     teamID.String(),
	})
	user, err := FindPaged(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.ManagedUser], error) {
			return r.client.User.GetAllManaged(ctx, po)
		},
		func(usr dtrack.ManagedUser) bool {
			return usr.Username == username
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read user",
			"Error with reading user: "+username+", in original error: "+err.Error(),
		)
		return
	}
	team, err := Find(user.Teams, func(team dtrack.Team) bool { return team.UUID == teamID })
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to identify user team membership",
			"Error with locating team: "+teamID.String()+", for user: "+username+", in original error: "+err.Error(),
		)
		return
	}

	state = userTeamResourceModel{
		Username: types.StringValue(user.Username),
		TeamID:   types.StringValue(team.UUID.String()),
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read User Team Membership", map[string]any{
		"username": state.Username.ValueString(),
		"team":     state.TeamID.ValueString(),
	})
}

func (*userTeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Nothing to Update. This resource only has Create, Delete actions.
	// Verifies that stored state is conformant to model.
	var plan userTeamResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated User Team Membership", map[string]any{
		"username": plan.Username.ValueString(),
		"team":     plan.TeamID.ValueString(),
	})
}

func (r *userTeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state userTeamResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, diag := TryParseUUID(state.TeamID, LifecycleDelete, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	username := state.Username.ValueString()

	tflog.Debug(ctx, "Deleting User Team Membership", map[string]any{
		"username": username,
		"team":     team.String(),
	})
	_, err := r.client.User.RemoveTeamFromUser(ctx, username, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete user team membership",
			"Error with username: "+username+", for team: "+team.String()+", in original error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted User Team Membership", map[string]any{
		"username": state.Username.ValueString(),
		"team":     state.TeamID.ValueString(),
	})
}

func (r *userTeamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
