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
	_ resource.Resource                = &aclMappingResource{}
	_ resource.ResourceWithConfigure   = &aclMappingResource{}
	_ resource.ResourceWithImportState = &aclMappingResource{}
)

func NewAclMappingResource() resource.Resource {
	return &aclMappingResource{}
}

type aclMappingResource struct {
	client *dtrack.Client
	semver *Semver
}

type aclMappingResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Team    types.String `tfsdk:"team"`
	Project types.String `tfsdk:"project"`
}

func (r *aclMappingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl_mapping"
}

func (r *aclMappingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an ACL mapping to grant a Team access to a Project",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID for the mapping. Has no meaning to DependencyTrack.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"team": schema.StringAttribute{
				Description: "UUID for the Team to whom to grant access.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project": schema.StringAttribute{
				Description: "UUID for the Project to which to grant access.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *aclMappingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan aclMappingResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	team, diag := TryParseUUID(plan.Team, LifecycleCreate, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	project, diag := TryParseUUID(plan.Project, LifecycleCreate, path.Root("project"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	mappingReq := dtrack.ACLMappingRequest{
		Team:    team,
		Project: project,
	}
	tflog.Debug(ctx, "Creating Project ACL Mapping", map[string]any{
		"project": mappingReq.Project.String(),
		"team":    mappingReq.Project.String(),
	})
	err := r.client.ACL.AddProjectMapping(ctx, mappingReq)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating acl mapping",
			"Unexpected error: "+err.Error(),
		)
		return
	}
	plan = aclMappingResourceModel{
		ID:      types.StringValue(fmt.Sprintf("%s/%s", team.String(), project.String())),
		Project: types.StringValue(project.String()),
		Team:    types.StringValue(team.String()),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created Project ACL Mapping", map[string]any{
		"id":      plan.ID.ValueString(),
		"project": plan.Project.ValueString(),
		"team":    plan.Project.ValueString(),
	})
}

func (r *aclMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state.
	var state aclMappingResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Refresh.
	team, diag := TryParseUUID(state.Team, LifecycleRead, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	projectId, diag := TryParseUUID(state.Project, LifecycleRead, path.Root("project"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Project ACL Mapping", map[string]any{
		"id":      state.ID.ValueString(),
		"team":    team.String(),
		"project": projectId.String(),
	})
	project, err := FindPaged(func(po dtrack.PageOptions) (dtrack.Page[dtrack.Project], error) {
		return r.client.ACL.GetAllProjects(ctx, team, po)
	}, func(project dtrack.Project) bool {
		return project.UUID == projectId
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get ACL mapping within Read",
			"Error with reading acl mapping for team: "+team.String()+", and project: "+projectId.String()+", in original error: "+err.Error(),
		)
		return
	}
	state = aclMappingResourceModel{
		ID:      types.StringValue(fmt.Sprintf("%s/%s", team.String(), project.UUID.String())),
		Team:    types.StringValue(team.String()),
		Project: types.StringValue(project.UUID.String()),
	}

	// Update state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Project ACL Mapping", map[string]any{
		"id":      state.ID.ValueString(),
		"project": state.Project.ValueString(),
		"team":    state.Team.ValueString(),
	})
}

func (r *aclMappingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Nothing to Update. This resource only has Create, Delete actions.
	// Get State.
	var plan aclMappingResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, diag := TryParseUUID(plan.Team, LifecycleUpdate, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	project, diag := TryParseUUID(plan.Project, LifecycleUpdate, path.Root("project"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	plan = aclMappingResourceModel{
		ID:      types.StringValue(fmt.Sprintf("%s/%s", team.String(), project.String())),
		Team:    types.StringValue(team.String()),
		Project: types.StringValue(project.String()),
	}

	// Update State.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated Project ACL Mapping", map[string]any{
		"id":      plan.ID.ValueString(),
		"project": plan.Project.ValueString(),
		"team":    plan.Team.ValueString(),
	})
}

func (r *aclMappingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state.
	var state aclMappingResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK.
	team, diag := TryParseUUID(state.Team, LifecycleDelete, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	project, diag := TryParseUUID(state.Project, LifecycleDelete, path.Root("project"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Execute.
	tflog.Debug(ctx, "Deleting Project ACL Mapping", map[string]any{
		"id":      state.ID.ValueString(),
		"team":    team.String(),
		"project": project.String(),
	})
	err := r.client.ACL.RemoveProjectMapping(ctx, team, project)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete acl mapping",
			"Unexpected error when trying to delete acl mapping, from error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted Project ACL Mapping", map[string]any{
		"id":      state.ID.ValueString(),
		"project": state.Project.ValueString(),
		"team":    state.Team.ValueString(),
	})
}

func (r *aclMappingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Within Import, unexpected id",
			"Expected id in format <Team UUID>/<Project UUID>. Received "+req.ID,
		)
		return
	}
	teamIdString := idParts[0]
	projectIdString := idParts[1]
	tflog.Debug(ctx, "Importing Project ACL Mapping", map[string]any{
		"team":    teamIdString,
		"project": projectIdString,
	})

	teamId, diag := TryParseUUID(types.StringValue(teamIdString), LifecycleImport, path.Root("team"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	projectId, diag := TryParseUUID(types.StringValue(projectIdString), LifecycleImport, path.Root("project"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	aclMappingState := aclMappingResourceModel{
		ID:      types.StringValue(fmt.Sprintf("%s/%s", teamId.String(), projectId.String())),
		Team:    types.StringValue(teamId.String()),
		Project: types.StringValue(projectId.String()),
	}
	diags := resp.State.Set(ctx, aclMappingState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported Project ACL Mapping", map[string]any{
		"id":      aclMappingState.ID.ValueString(),
		"project": aclMappingState.Project.ValueString(),
		"team":    aclMappingState.Team.ValueString(),
	})
}

func (r *aclMappingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
