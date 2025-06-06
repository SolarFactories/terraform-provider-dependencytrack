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
	_ resource.Resource                = &policyResource{}
	_ resource.ResourceWithConfigure   = &policyResource{}
	_ resource.ResourceWithImportState = &policyResource{}
)

type (
	policyResource struct {
		client *dtrack.Client
		semver *Semver
	}

	policyResourceModel struct {
		ID        types.String `tfsdk:"id"`
		Name      types.String `tfsdk:"name"`
		Operator  types.String `tfsdk:"operator"`
		Violation types.String `tfsdk:"violation"`
	}
)

func NewPolicyResource() resource.Resource {
	return &policyResource{}
}

func (*policyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (*policyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID for the Policy as generated by DependencyTrack.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Policy.",
				Required:    true,
			},
			"operator": schema.StringAttribute{
				Description: "Operator to apply to conditions. See DependencyTrack for allowed values.",
				Required:    true,
			},
			"violation": schema.StringAttribute{
				Description: "Violation state for when a condition fails. See DependencyTrack for allowed values.",
				Required:    true,
			},
		},
	}
}

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan policyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyReq := dtrack.Policy{
		Name:           plan.Name.ValueString(),
		Operator:       dtrack.PolicyOperator(plan.Operator.ValueString()),
		ViolationState: dtrack.PolicyViolationState(plan.Violation.ValueString()),
	}
	tflog.Debug(ctx, "Creating Policy", map[string]any{
		"name":      policyReq.Name,
		"operator":  string(policyReq.Operator),
		"violation": string(policyReq.ViolationState),
	})
	policyRes, err := r.client.Policy.Create(ctx, policyReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating policy",
			"Error from: "+err.Error(),
		)
		return
	}
	plan = policyResourceModel{
		ID:        types.StringValue(policyRes.UUID.String()),
		Name:      types.StringValue(policyRes.Name),
		Operator:  types.StringValue(string(policyRes.Operator)),
		Violation: types.StringValue(string(policyRes.ViolationState)),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created Policy", map[string]any{
		"id":        plan.ID.ValueString(),
		"name":      plan.Name.ValueString(),
		"operator":  plan.Operator.ValueString(),
		"violation": plan.Violation.ValueString(),
	})
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state.
	var state policyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refresh.
	id, diag := TryParseUUID(state.ID, LifecycleRead, path.Root("id"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	tflog.Debug(ctx, "Reading Policy", map[string]any{
		"id":        id.String(),
		"name":      state.Name.ValueString(),
		"operator":  state.Operator.ValueString(),
		"violation": state.Violation.ValueString(),
	})

	policy, err := r.client.Policy.Get(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get updated policy",
			"Error with reading policy: "+id.String()+", from: "+err.Error(),
		)
		return
	}
	state = policyResourceModel{
		ID:        types.StringValue(policy.UUID.String()),
		Name:      types.StringValue(policy.Name),
		Operator:  types.StringValue(string(policy.Operator)),
		Violation: types.StringValue(string(policy.ViolationState)),
	}

	// Update state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Policy", map[string]any{
		"id":        state.ID.ValueString(),
		"name":      state.Name.ValueString(),
		"operator":  state.Operator.ValueString(),
		"violation": state.Violation.ValueString(),
	})
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get State.
	var plan policyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK.
	id, diag := TryParseUUID(plan.ID, LifecycleUpdate, path.Root("id"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	policyReq := dtrack.Policy{
		UUID:           id,
		Name:           plan.Name.ValueString(),
		Operator:       dtrack.PolicyOperator(plan.Operator.ValueString()),
		ViolationState: dtrack.PolicyViolationState(plan.Violation.ValueString()),
	}
	// Execute.
	tflog.Debug(ctx, "Updating Policy", map[string]any{
		"id":        policyReq.UUID.String(),
		"name":      policyReq.Name,
		"operator":  policyReq.Operator,
		"violation": policyReq.ViolationState,
	})
	policyRes, err := r.client.Policy.Update(ctx, policyReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update policy",
			"Error in: "+id.String()+", from: "+err.Error(),
		)
		return
	}

	// Map SDK to TF.
	plan = policyResourceModel{
		ID:        types.StringValue(policyRes.UUID.String()),
		Name:      types.StringValue(policyRes.Name),
		Operator:  types.StringValue(string(policyRes.Operator)),
		Violation: types.StringValue(string(policyRes.ViolationState)),
	}

	// Update State.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated Policy", map[string]any{
		"id":        plan.ID.ValueString(),
		"name":      plan.Name.ValueString(),
		"operator":  plan.Operator.ValueString(),
		"violation": plan.Violation.ValueString(),
	})
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state.
	var state policyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK.
	id, diag := TryParseUUID(state.ID, LifecycleDelete, path.Root("id"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	// Execute.
	tflog.Debug(ctx, "Deleting Policy", map[string]any{
		"id":        id.String(),
		"name":      state.Name.ValueString(),
		"operator":  state.Operator.ValueString(),
		"violation": state.Violation.ValueString(),
	})
	err := r.client.Policy.Delete(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete policy",
			"Unexpected error when trying to delete policy: "+id.String()+", from error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted Policy", map[string]any{
		"id":        state.ID.ValueString(),
		"name":      state.Name.ValueString(),
		"operator":  state.Operator.ValueString(),
		"violation": state.Violation.ValueString(),
	})
}

func (*policyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing Policy", map[string]any{
		"id": req.ID,
	})
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported Policy", map[string]any{
		"id": req.ID,
	})
}

func (r *policyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
