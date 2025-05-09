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
	_ resource.Resource              = &policyTagResource{}
	_ resource.ResourceWithConfigure = &policyTagResource{}
)

type (
	policyTagResource struct {
		client *dtrack.Client
		semver *Semver
	}

	policyTagResourceModel struct {
		PolicyID types.String `tfsdk:"policy"`
		Tag      types.String `tfsdk:"tag"`
	}
)

func NewPolicyTagResource() resource.Resource {
	return &policyTagResource{}
}

// TODO: Once have `dependencytrack_tag` resource to create tags, add testing.
func (*policyTagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_tag"
}

func (*policyTagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an application of a Policy to a Tag.",
		Attributes: map[string]schema.Attribute{
			"policy": schema.StringAttribute{
				Description: "UUID for the Policy to apply to the Tag.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tag": schema.StringAttribute{
				Description: "Name of the Tag to which to apply Policy.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *policyTagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan policyTagResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	policyID, diag := TryParseUUID(plan.PolicyID, LifecycleCreate, path.Root("policy"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	tagName := plan.Tag.ValueString()

	tflog.Debug(ctx, "Creating Policy Tag Mapping", map[string]any{
		"policy": policyID.String(),
		"tag":    tagName,
	})
	policy, err := r.client.Policy.AddTag(ctx, policyID, tagName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Policy Tag Mapping",
			"Error from: "+err.Error(),
		)
		return
	}
	plan = policyTagResourceModel{
		PolicyID: types.StringValue(policy.UUID.String()),
		Tag:      types.StringValue(tagName),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created Policy Tag Mapping", map[string]any{
		"policy": plan.PolicyID.ValueString(),
		"tag":    plan.Tag.ValueString(),
	})
}

func (r *policyTagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state.
	var state policyTagResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refresh.
	policyID, diag := TryParseUUID(state.PolicyID, LifecycleRead, path.Root("policy"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	tagName := state.Tag.ValueString()

	tflog.Debug(ctx, "Reading Policy Tag Mapping", map[string]any{
		"policy": policyID.String(),
		"tag":    tagName,
	})
	policy, err := r.client.Policy.Get(ctx, policyID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to retrieve policy",
			"Error from: "+err.Error(),
		)
		return
	}
	tag, err := Find(policy.Tags, func(tag dtrack.Tag) bool {
		return tag.Name == tagName
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to locate Policy Tag Mapping",
			"Error from: "+err.Error(),
		)
		return
	}
	state = policyTagResourceModel{
		PolicyID: types.StringValue(policy.UUID.String()),
		Tag:      types.StringValue(tag.Name),
	}

	// Update state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Policy Tag Mapping", map[string]any{
		"policy": state.PolicyID.ValueString(),
		"tag":    state.Tag.ValueString(),
	})
}

func (*policyTagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Resource has nothing to update, as it bridges by it's existence. Existence check is done within `Read`.
	// Get State.
	var plan policyTagResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	policyID, diag := TryParseUUID(plan.PolicyID, LifecycleUpdate, path.Root("policy"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	tagName := plan.Tag.ValueString()
	tflog.Debug(ctx, "Updating Policy Tag Mapping", map[string]any{
		"policy": policyID.String(),
		"tag":    tagName,
	})

	plan = policyTagResourceModel{
		PolicyID: types.StringValue(policyID.String()),
		Tag:      types.StringValue(tagName),
	}

	// Update State.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated Policy Tag Mapping", map[string]any{
		"policy": plan.PolicyID.ValueString(),
		"tag":    plan.Tag.ValueString(),
	})
}

func (r *policyTagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state.
	var state policyTagResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	policyID, diag := TryParseUUID(state.PolicyID, LifecycleDelete, path.Root("policy"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	tagName := state.Tag.ValueString()

	tflog.Debug(ctx, "Deleting Policy Tag Mapping", map[string]any{
		"policy": policyID.String(),
		"tag":    tagName,
	})
	_, err := r.client.Policy.DeleteTag(ctx, policyID, tagName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete Policy Tag Mapping",
			"Error from: "+err.Error(),
		)
	}
	tflog.Debug(ctx, "Deleted Policy Tag Mapping", map[string]any{
		"policy": state.PolicyID.ValueString(),
		"tag":    state.Tag.ValueString(),
	})
}

func (r *policyTagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
