package provider

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource                = &notificationRuleTeamResource{}
	_ resource.ResourceWithConfigure   = &notificationRuleTeamResource{}
	_ resource.ResourceWithImportState = &notificationRuleTeamResource{}
)

type (
	notificationRuleTeamResource struct {
		client *dtrack.Client
		semver *Semver
	}

	notificationRuleTeamResourceModel struct {
		ID   types.String `tfsdk:"id"`
		Rule types.String `tfsdk:"rule"`
		Team types.String `tfsdk:"team"`
	}
)

func NewNotificationRuleTeamResource() resource.Resource {
	return &notificationRuleTeamResource{}
}

func (*notificationRuleTeamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_rule_team"
}

func (*notificationRuleTeamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a mapping for a Notification Rule to a Team",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID for the mapping. Has no meaning to DependencyTrack.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"rule": schema.StringAttribute{
				Description: "UUID for the Notification Rule to set for Team. Must use an email publisher.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team": schema.StringAttribute{
				Description: "UUID for the Team for which to generate Notifications.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *notificationRuleTeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationRuleTeamResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleID, diag := TryParseUUID(plan.Rule, LifecycleCreate, path.Root("rule"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	teamID, diag := TryParseUUID(plan.Team, LifecycleCreate, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating Notification Rule Team Mapping", map[string]any{
		"rule": ruleID.String(),
		"team": teamID.String(),
	})

	_, err := r.client.Notification.AddTeamToRule(ctx, ruleID, teamID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create assignment of Notification Rule to Team",
			"Error with rule with id: "+ruleID.String()+", and team with id: "+teamID.String()+", in original error: "+err.Error(),
		)
		return
	}

	newState := notificationRuleTeamResourceModel{
		ID:   types.StringValue(fmt.Sprintf("%s/%s", ruleID.String(), teamID.String())),
		Team: types.StringValue(teamID.String()),
		Rule: types.StringValue(ruleID.String()),
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created Notification Rule Team Mapping", map[string]any{
		"id":   newState.ID.ValueString(),
		"rule": newState.Rule.ValueString(),
		"team": newState.Team.ValueString(),
	})
}

func (r *notificationRuleTeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state.
	var state notificationRuleTeamResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Refresh.
	ruleID, diag := TryParseUUID(state.Rule, LifecycleRead, path.Root("rule"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	teamID, diag := TryParseUUID(state.Team, LifecycleRead, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Notification Rule Team Mapping", map[string]any{
		"id":   state.ID.ValueString(),
		"rule": ruleID.String(),
		"team": teamID.String(),
	})
	rule, err := FindPaged(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.NotificationRule], error) {
			return r.client.Notification.GetAllRules(ctx, po, dtrack.SortOptions{}, dtrack.GetAllRulesFilterOptions{})
		},
		func(rule dtrack.NotificationRule) bool {
			return rule.UUID == ruleID
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Notification Rule Team Mapping",
			"Error for rule with id: "+ruleID.String()+", and team with id: "+teamID.String()+", in original error: "+err.Error(),
		)
		return
	}
	team, err := Find(rule.Teams, func(team dtrack.Team) bool {
		return team.UUID == teamID
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to locate Notification Rule Team Mapping when reading",
			"Error for rule with id: "+ruleID.String()+", and team with id: "+teamID.String()+", in original error: "+err.Error(),
		)
		return
	}

	state = notificationRuleTeamResourceModel{
		ID:   types.StringValue(fmt.Sprintf("%s/%s", rule.UUID.String(), team.UUID.String())),
		Rule: types.StringValue(rule.UUID.String()),
		Team: types.StringValue(team.UUID.String()),
	}

	// Update state.
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Notification Rule Team Mapping", map[string]any{
		"id":   state.ID.ValueString(),
		"rule": state.Rule.ValueString(),
		"team": state.Team.ValueString(),
	})
}

func (*notificationRuleTeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Nothing to Update. This resource only has Create, Delete actions.
	// Get State.
	var plan notificationRuleTeamResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating Notification Rule Team Mapping", map[string]any{
		"id":   plan.ID.ValueString(),
		"rule": plan.Rule.ValueString(),
		"team": plan.Team.ValueString(),
	})

	ruleID, diag := TryParseUUID(plan.Rule, LifecycleUpdate, path.Root("rule"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	teamID, diag := TryParseUUID(plan.Team, LifecycleUpdate, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	newState := notificationRuleTeamResourceModel{
		ID:   types.StringValue(fmt.Sprintf("%s/%s", ruleID.String(), teamID.String())),
		Rule: types.StringValue(ruleID.String()),
		Team: types.StringValue(teamID.String()),
	}

	// Update State.
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated Notification Rule Team Mapping", map[string]any{
		"id":   newState.ID.ValueString(),
		"rule": newState.Rule.ValueString(),
		"team": newState.Team.ValueString(),
	})
}

func (r *notificationRuleTeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state.
	var state notificationRuleTeamResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK.
	ruleID, diag := TryParseUUID(state.Rule, LifecycleDelete, path.Root("rule"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	teamID, diag := TryParseUUID(state.Team, LifecycleDelete, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Execute.
	tflog.Debug(ctx, "Deleting Notification Rule Team Mapping", map[string]any{
		"id":   state.ID.ValueString(),
		"rule": ruleID.String(),
		"team": teamID.String(),
	})
	_, err := r.client.Notification.RemoveTeamFromRule(ctx, ruleID, teamID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Notification Rule Team Mapping",
			"Error for rule with id: "+ruleID.String()+", and team with id: "+teamID.String()+", in original error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted Notification Rule Team Mapping", map[string]any{
		"id":   state.ID.ValueString(),
		"rule": state.Rule.ValueString(),
		"team": state.Team.ValueString(),
	})
}

func (*notificationRuleTeamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Within Import, unexpected id",
			"Expected id in format <Rule UUID>/<Team UUID>. Received "+req.ID,
		)
		return
	}
	ruleIDString := idParts[0]
	teamIDString := idParts[1]
	tflog.Debug(ctx, "Importing Notification Rule Team Mapping", map[string]any{
		"rule": ruleIDString,
		"team": teamIDString,
	})

	ruleID, diag := TryParseUUID(types.StringValue(ruleIDString), LifecycleImport, path.Root("rule"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	teamID, diag := TryParseUUID(types.StringValue(teamIDString), LifecycleImport, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	newState := notificationRuleTeamResourceModel{
		ID:   types.StringValue(fmt.Sprintf("%s/%s", ruleID.String(), teamID.String())),
		Rule: types.StringValue(ruleID.String()),
		Team: types.StringValue(teamID.String()),
	}
	diags := resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported Notification Rule Team Mapping", map[string]any{
		"id":   newState.ID.ValueString(),
		"rule": newState.Rule.ValueString(),
		"team": newState.Team.ValueString(),
	})
}

func (r *notificationRuleTeamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
