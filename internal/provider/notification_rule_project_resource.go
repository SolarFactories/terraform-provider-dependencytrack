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
	_ resource.Resource                = &notificationRuleProjectResource{}
	_ resource.ResourceWithConfigure   = &notificationRuleProjectResource{}
	_ resource.ResourceWithImportState = &notificationRuleProjectResource{}
)

type (
	notificationRuleProjectResource struct {
		client *dtrack.Client
		semver *Semver
	}

	notificationRuleProjectResourceModel struct {
		ID      types.String `tfsdk:"id"`
		Rule    types.String `tfsdk:"rule"`
		Project types.String `tfsdk:"project"`
	}
)

func NewNotificationRuleProjectResource() resource.Resource {
	return &notificationRuleProjectResource{}
}

func (*notificationRuleProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_rule_project"
}

func (*notificationRuleProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a mapping for a Notification Rule to a Project",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID for the mapping. Has no meaning to DependencyTrack.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"rule": schema.StringAttribute{
				Description: "UUID for the Notification Rule to set for Project.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project": schema.StringAttribute{
				Description: "UUID for the Project for which to generate Notifications.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *notificationRuleProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationRuleProjectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleID, diag := TryParseUUID(plan.Rule, LifecycleCreate, path.Root("rule"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	projectID, diag := TryParseUUID(plan.Project, LifecycleCreate, path.Root("project"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating Notification Rule Project Mapping", map[string]any{
		"rule":    ruleID.String(),
		"project": projectID.String(),
	})

	_, err := r.client.Notification.AddProjectToRule(ctx, ruleID, projectID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create assignment of Notification Rule to Project",
			"Error with rule with id: "+ruleID.String()+", and project with id: "+projectID.String()+", in original error: "+err.Error(),
		)
		return
	}

	newState := notificationRuleProjectResourceModel{
		ID:      types.StringValue(fmt.Sprintf("%s/%s", ruleID.String(), projectID.String())),
		Project: types.StringValue(projectID.String()),
		Rule:    types.StringValue(ruleID.String()),
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created Notification Rule Project Mapping", map[string]any{
		"id":      newState.ID.ValueString(),
		"rule":    newState.Rule.ValueString(),
		"project": newState.Project.ValueString(),
	})
}

func (r *notificationRuleProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state.
	var state notificationRuleProjectResourceModel
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
	projectID, diag := TryParseUUID(state.Project, LifecycleRead, path.Root("project"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Notification Rule Project Mapping", map[string]any{
		"id":      state.ID.ValueString(),
		"rule":    ruleID.String(),
		"project": projectID.String(),
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
			"Unable to read Notification Rule Project Mapping",
			"Error for rule with id: "+ruleID.String()+", and project with id: "+projectID.String()+", in original error: "+err.Error(),
		)
		return
	}
	project, err := Find(rule.Projects, func(project dtrack.Project) bool {
		return project.UUID == projectID
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to locate Notification Rule Project Mapping when reading",
			"Error for rule with id: "+ruleID.String()+", and project with id: "+projectID.String()+", in original error: "+err.Error(),
		)
		return
	}

	state = notificationRuleProjectResourceModel{
		ID:      types.StringValue(fmt.Sprintf("%s/%s", rule.UUID.String(), project.UUID.String())),
		Rule:    types.StringValue(rule.UUID.String()),
		Project: types.StringValue(project.UUID.String()),
	}

	// Update state.
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Notification Rule Project Mapping", map[string]any{
		"id":      state.ID.ValueString(),
		"rule":    state.Rule.ValueString(),
		"project": state.Project.ValueString(),
	})
}

func (*notificationRuleProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Nothing to Update. This resource only has Create, Delete actions.
	// Get State.
	var plan notificationRuleProjectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating Notification Rule Project Mapping", map[string]any{
		"id":      plan.ID.ValueString(),
		"rule":    plan.Rule.ValueString(),
		"project": plan.Project.ValueString(),
	})

	ruleID, diag := TryParseUUID(plan.Rule, LifecycleUpdate, path.Root("rule"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	projectID, diag := TryParseUUID(plan.Project, LifecycleUpdate, path.Root("project"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	newState := notificationRuleProjectResourceModel{
		ID:      types.StringValue(fmt.Sprintf("%s/%s", ruleID.String(), projectID.String())),
		Rule:    types.StringValue(ruleID.String()),
		Project: types.StringValue(projectID.String()),
	}

	// Update State.
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated Notification Rule Project Mapping", map[string]any{
		"id":      plan.ID.ValueString(),
		"rule":    plan.Rule.ValueString(),
		"project": plan.Project.ValueString(),
	})
}

func (r *notificationRuleProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state.
	var state notificationRuleProjectResourceModel
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
	projectID, diag := TryParseUUID(state.Project, LifecycleDelete, path.Root("project"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Execute.
	tflog.Debug(ctx, "Deleting Notification Rule Project Mapping", map[string]any{
		"id":      state.ID.ValueString(),
		"rule":    ruleID.String(),
		"project": projectID.String(),
	})
	_, err := r.client.Notification.RemoveProjectFromRule(ctx, ruleID, projectID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Notification Rule Project Mapping",
			"Error for rule with id: "+ruleID.String()+", and project with id: "+projectID.String()+", in orifinal error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted Notification Rule Project Mapping", map[string]any{
		"id":      state.ID.ValueString(),
		"rule":    state.Rule.ValueString(),
		"project": state.Project.ValueString(),
	})
}

func (*notificationRuleProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Within Import, unexpected id",
			"Expected id in format <Rule UUID>/<Project UUID>. Received "+req.ID,
		)
		return
	}
	ruleIDString := idParts[0]
	projectIDString := idParts[1]
	tflog.Debug(ctx, "Importing Notification Rule Project Mapping", map[string]any{
		"rule":    ruleIDString,
		"project": projectIDString,
	})

	ruleID, diag := TryParseUUID(types.StringValue(ruleIDString), LifecycleImport, path.Root("rule"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	projectID, diag := TryParseUUID(types.StringValue(projectIDString), LifecycleImport, path.Root("project"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	newState := notificationRuleProjectResourceModel{
		ID:      types.StringValue(fmt.Sprintf("%s/%s", ruleID.String(), projectID.String())),
		Rule:    types.StringValue(ruleID.String()),
		Project: types.StringValue(projectID.String()),
	}
	diags := resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported Notification Rule Project Mapping", map[string]any{
		"id":      newState.ID.ValueString(),
		"rule":    newState.Rule.ValueString(),
		"project": newState.Project.ValueString(),
	})
}

func (r *notificationRuleProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
