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
	_ resource.Resource              = &policyTagResource{}
	_ resource.ResourceWithConfigure = &policyTagResource{}
)

func NewPolicyTagResource() resource.Resource {
	return &policyTagResource{}
}

type policyTagResource struct {
	client *dtrack.Client
	semver *Semver
}

type policyTagResourceModel struct {
	PolicyID types.String `tfsdk:"policy"`
	Tag      types.String `tfsdk:"tag"`
}

// TODO: Once have `dependencytrack_tag` resource to create tags, add testing.
func (r *policyTagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_tag"
}

func (r *policyTagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
	policyId, err := uuid.Parse(plan.PolicyID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("policy"),
			"Within Create, unable to parse policy into UUID",
			"Error from: "+err.Error(),
		)
		return
	}
	tagName := plan.Tag.ValueString()

	tflog.Debug(ctx, "Adding policy with id: "+policyId.String()+" to tag: "+tagName)
	policy, err := r.client.Policy.AddTag(ctx, policyId, tagName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating tag policy mapping",
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
	tflog.Debug(ctx, "Created policy-tag mapping for policy with id: "+plan.PolicyID.ValueString()+" to tag with name: "+plan.Tag.ValueString())
}

func (r *policyTagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state
	var state policyTagResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Refreshing policy-tag mapping for policy: "+state.PolicyID.ValueString()+", and tag: "+state.Tag.ValueString())

	// Refresh
	policyId, err := uuid.Parse(state.PolicyID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("policy"),
			"Within Read, unable to parse policy into UUID",
			"Error from: "+err.Error(),
		)
		return
	}
	tagName := state.Tag.ValueString()

	policy, err := r.client.Policy.Get(ctx, policyId)
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
			"Within Read, unable to locate tag-policy mapping",
			"Error from: "+err.Error(),
		)
		return
	}
	state = policyTagResourceModel{
		PolicyID: types.StringValue(policy.UUID.String()),
		Tag:      types.StringValue(tag.Name),
	}

	// Update state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Refreshed tag-policy mapping for policy: "+state.PolicyID.ValueString()+", with tag: "+state.Tag.ValueString())
}

func (r *policyTagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Resource has nothing to update, as it bridges by it's existence. Existence check is done within `Read`.
	// Get State
	var plan policyTagResourceModel
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
		return
	}
	tagName := plan.Tag.ValueString()

	plan = policyTagResourceModel{
		PolicyID: types.StringValue(policyId.String()),
		Tag:      types.StringValue(tagName),
	}

	// Update State
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated tag-policy mapping for policy: "+plan.PolicyID.ValueString()+", and tag: "+plan.Tag.ValueString())
}

func (r *policyTagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state
	var state policyTagResourceModel
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
		return
	}
	tagName := state.Tag.ValueString()

	tflog.Debug(ctx, "Deleting tag-policy mapping for policy: "+policyId.String()+", with tag: "+tagName)
	_, err = r.client.Policy.DeleteTag(ctx, policyId, tagName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete tag-policy mapping",
			"Error from: "+err.Error(),
		)
	}
	tflog.Debug(ctx, "Deleted tag-policy mapping for policy: "+policyId.String()+", with tag: "+tagName)
}

func (r *policyTagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
