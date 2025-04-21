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
	_ resource.Resource              = &policyProjectResource{}
	_ resource.ResourceWithConfigure = &policyProjectResource{}
)

func NewPolicyProjectResource() resource.Resource {
	return &policyProjectResource{}
}

type policyProjectResource struct {
	client *dtrack.Client
	semver *Semver
}

type policyProjectResourceModel struct {
	PolicyID  types.String `tfsdk:"policy"`
	ProjectID types.String `tfsdk:"project"`
}

func (r *policyProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_project"
}

func (r *policyProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an application of a Policy to a Project.",
		Attributes: map[string]schema.Attribute{
			"policy": schema.StringAttribute{
				Description: "UUID for the Policy to apply to the Project.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project": schema.StringAttribute{
				Description: "UUID for the Project for which to apply Policy.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *policyProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan policyProjectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	policyId, err := uuid.Parse(plan.PolicyID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("policy"),
			"Within Create, unable to parse policy into UUID",
			"Error from: "+err.Error(),
		)
	}
	projectId, err := uuid.Parse(plan.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("project"),
			"Within Create, unable to parse project into UUID",
			"Error from: "+err.Error(),
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Adding policy-project for policy with id: "+policyId.String()+" to project: "+projectId.String())
	policy, err := r.client.Policy.AddProject(ctx, policyId, projectId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project policy mapping",
			"Unexpected error: "+err.Error(),
		)
		return
	}
	plan = policyProjectResourceModel{
		PolicyID:  types.StringValue(policy.UUID.String()),
		ProjectID: types.StringValue(projectId.String()),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created policy-project mapping for policy with id: "+plan.PolicyID.ValueString()+" to project with id: "+plan.ProjectID.ValueString())
}

func (r *policyProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state
	var state policyProjectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Refreshing policy-project mapping for policy: "+state.PolicyID.ValueString()+", and project: "+state.ProjectID.ValueString())

	// Refresh
	policyId, err := uuid.Parse(state.PolicyID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("policy"),
			"Within Read, unable to parse policy into UUID",
			"Error from: "+err.Error(),
		)
	}
	projectId, err := uuid.Parse(state.ProjectID.ValueString())
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

	policy, err := r.client.Policy.Get(ctx, policyId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to retrieve policy",
			"Error from: "+err.Error(),
		)
		return
	}
	project, err := Find(policy.Projects, func(project dtrack.Project) bool {
		return project.UUID == projectId
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to locate project-policy mapping",
			"Error from: "+err.Error(),
		)
		return
	}
	state = policyProjectResourceModel{
		PolicyID:  types.StringValue(policy.UUID.String()),
		ProjectID: types.StringValue(project.UUID.String()),
	}

	// Update state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Refreshed project-policy mapping for policy: "+state.PolicyID.ValueString()+", with project: "+state.ProjectID.ValueString())
}

func (r *policyProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Resource has nothing to update, as it bridges by it's existence. Existence check is done within `Read`.
	// Get State
	var plan policyProjectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	policyId, err := uuid.Parse(plan.PolicyID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("policy"),
			"Within Update, unable to parse policy into UUID",
			"Error from: "+err.Error(),
		)
	}
	projectId, err := uuid.Parse(plan.ProjectID.ValueString())
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

	plan = policyProjectResourceModel{
		PolicyID:  types.StringValue(policyId.String()),
		ProjectID: types.StringValue(projectId.String()),
	}

	// Update State
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated project-policy mapping for policy: "+plan.PolicyID.ValueString()+", and project: "+plan.ProjectID.ValueString())
}

func (r *policyProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state
	var state policyProjectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	policyId, err := uuid.Parse(state.PolicyID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("policy"),
			"Within Delete, unable to parse policy into UUID",
			"Error from: "+err.Error(),
		)
	}
	projectId, err := uuid.Parse(state.ProjectID.ValueString())
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

	tflog.Debug(ctx, "Deleting project-policy mapping for policy: "+policyId.String()+", with project: "+projectId.String())
	_, err = r.client.Policy.DeleteProject(ctx, policyId, projectId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete project-policy mapping",
			"Error from: "+err.Error(),
		)
	}
	tflog.Debug(ctx, "Deleted project-policy mapping for policy: "+policyId.String()+", with project: "+projectId.String())
}

func (r *policyProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
