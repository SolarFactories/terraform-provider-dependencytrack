package provider

import (
	"context"
	"fmt"
	"strings"

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
			"group": schema.StringAttribute{
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
	team, err := uuid.Parse(plan.Team.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("team"),
			"Within Create, unable to parse team into UUID",
			"Error from: "+err.Error(),
		)
		return
	}
	project, err := uuid.Parse(plan.Project.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("project"),
			"Within Create, unable to parse project into UUID",
			"Error from: "+err.Error(),
		)
		return
	}

	mappingReq := dtrack.ACLMappingRequest{
		Team:    team,
		Project: project,
	}
	tflog.Debug(ctx, "Granting ACL for project "+mappingReq.Project.String()+" to team "+mappingReq.Team.String())
	err = r.client.ACLService.AddProjectMapping(ctx, mappingReq)

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
	tflog.Debug(ctx, "Created acl for project with id: "+plan.Project.ValueString()+" to team, with id: "+plan.Team.ValueString())
}

func (r *aclMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state
	var state aclMappingResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Refreshing acl mapping for team: "+state.Team.ValueString()+", and project: "+state.Project.ValueString())

	// Refresh
	team, err := uuid.Parse(state.Team.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("team"),
			"Within Read, unable to parse team into UUID",
			"Error from: "+err.Error(),
		)
	}
	projectId, err := uuid.Parse(state.Project.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("project"),
			"Within Read, unable to parse project into UUID",
			"Error from: "+err.Error(),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}
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

	// Update state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Refreshed acl mapping for team: "+state.Team.ValueString()+", and project: "+state.Project.ValueString())
}

func (r *aclMappingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Nothing to Update. This resource only has Create, Delete actions
	// Get State
	var plan aclMappingResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := uuid.Parse(plan.Team.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("team"),
			"Within Update, unable to parse team into UUID",
			"Error from: "+err.Error(),
		)
	}
	project, err := uuid.Parse(plan.Project.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("project"),
			"Within Update, unable to parse project into UUID",
			"Error from: "+err.Error(),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	plan = aclMappingResourceModel{
		ID:      types.StringValue(fmt.Sprintf("%s/%s", team.String(), project.String())),
		Team:    types.StringValue(team.String()),
		Project: types.StringValue(project.String()),
	}

	// Update State
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated acl mapping for team: "+plan.Team.ValueString()+", and project: "+plan.Project.ValueString())
}

func (r *aclMappingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state
	var state aclMappingResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK
	team, err := uuid.Parse(state.Team.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("team"),
			"Within Delete, unable to parse team into UUID",
			"Error from: "+err.Error(),
		)
	}
	project, err := uuid.Parse(state.Project.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("project"),
			"Within Delete, unable to parse project into UUID",
			"Error from: "+err.Error(),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Execute
	tflog.Debug(ctx, "Deleting acl mapping with for project with id: "+state.Project.ValueString()+", and team with id: "+state.Team.ValueString())
	err = r.client.ACL.RemoveProjectMapping(ctx, team, project)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete acl mapping",
			"Unexpected error when trying to delete acl mapping, from error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted acl group mapping with id: "+state.ID.ValueString())
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

	teamId, err := uuid.Parse(teamIdString)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Import, unable to parse team into UUID",
			"Error from: "+err.Error(),
		)
	}
	projectId, err := uuid.Parse(projectIdString)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Import, unable to parse project into UUID",
			"Error from: "+err.Error(),
		)
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
	tflog.Debug(ctx, "Imported an ACL Mapping from team: "+aclMappingState.Team.ValueString()+", and project: "+aclMappingState.Project.ValueString())

}

func (r *aclMappingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
