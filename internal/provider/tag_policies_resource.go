package provider

import (
	"context"
	"fmt"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &tagPoliciesResource{}
	_ resource.ResourceWithConfigure   = &tagPoliciesResource{}
	_ resource.ResourceWithImportState = &tagPoliciesResource{}
)

type (
	tagPoliciesResource struct {
		client *dtrack.Client
		semver *Semver
	}

	tagPoliciesResourceModel struct {
		ID       types.String   `tfsdk:"id"`
		Tag      types.String   `tfsdk:"tag"`
		Policies []types.String `tfsdk:"policies"`
	}
)

func NewTagPoliciesResource() resource.Resource {
	return &tagPoliciesResource{}
}

func (*tagPoliciesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag_policies"
}

func (*tagPoliciesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Applies a tag to multiple policies. Requires API version >= 4.12.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Name of the Tag.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tag": schema.StringAttribute{
				Description: "Name of the Tag.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policies": schema.ListAttribute{
				Description: "Policy UUIDs to which to apply tag. Will present a delta, unless sorted by policy name.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *tagPoliciesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tagPoliciesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagName := plan.Tag.ValueString()
	currentPoliciesInfo, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.TaggedPolicyListResponseItem], error) {
		return r.client.Tag.GetPolicies(ctx, tagName, po, dtrack.SortOptions{})
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Create, unable to request current policies for tag: "+tagName,
			"Error from: "+err.Error(),
		)
		return
	}
	currentPolicies := Map(currentPoliciesInfo, func(current dtrack.TaggedPolicyListResponseItem) uuid.UUID { return current.UUID })
	desiredPolicies, err := TryMap(plan.Policies, func(value types.String) (uuid.UUID, error) {
		return uuid.Parse(value.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Create, unable to parse policy into UUID",
			"Error from: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Creating Tag Policies", map[string]any{
		"tag":     tagName,
		"current": currentPolicies,
		"desired": desiredPolicies,
	})

	addPolicies, removePolicies := ListDeltasUUID(currentPolicies, desiredPolicies)
	if len(addPolicies) > 0 {
		err = r.client.Tag.TagPolicies(ctx, tagName, addPolicies)
		if err != nil {
			resp.Diagnostics.AddError(
				"Within Create, unable to add policies from tag list for tag: "+tagName,
				"Error from: "+err.Error(),
			)
		}
	}
	if len(removePolicies) > 0 {
		err = r.client.Tag.UntagPolicies(ctx, tagName, removePolicies)
		if err != nil {
			resp.Diagnostics.AddError(
				"Within Create, unable to remove policies from tag list for tag: "+tagName,
				"Error from: "+err.Error(),
			)
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	plan = tagPoliciesResourceModel{
		ID:  types.StringValue(tagName),
		Tag: types.StringValue(tagName),
		Policies: Map(desiredPolicies, func(policy uuid.UUID) types.String {
			return types.StringValue(policy.String())
		}),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created Tag Policies", map[string]any{
		"tag":      plan.Tag.ValueString(),
		"policies": desiredPolicies,
	})
}

func (r *tagPoliciesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tagPoliciesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagName := state.ID.ValueString()
	tflog.Debug(ctx, "Reading Tag Policies", map[string]any{
		"tag":        tagName,
		"policies.#": len(state.Policies),
	})

	taggedPoliciesInfo, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.TaggedPolicyListResponseItem], error) {
		return r.client.Tag.GetPolicies(ctx, tagName, po, dtrack.SortOptions{})
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to get current list of policies for tag: "+tagName,
			"Error from: "+err.Error(),
		)
		return
	}

	state = tagPoliciesResourceModel{
		ID:  types.StringValue(tagName),
		Tag: types.StringValue(tagName),
		Policies: Map(taggedPoliciesInfo, func(info dtrack.TaggedPolicyListResponseItem) types.String {
			return types.StringValue(info.UUID.String())
		}),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Tag Policies", map[string]any{
		"tag":      state.Tag.ValueString(),
		"policies": Map(state.Policies, func(v types.String) string { return v.ValueString() }),
	})
}

func (r *tagPoliciesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tagPoliciesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagName := plan.Tag.ValueString()
	tflog.Debug(ctx, "Updating Tag Policies", map[string]any{
		"tag":        tagName,
		"policies.#": len(plan.Policies),
	})

	currentPoliciesInfo, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.TaggedPolicyListResponseItem], error) {
		return r.client.Tag.GetPolicies(ctx, tagName, po, dtrack.SortOptions{})
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Update, unable to request current policies for tag: "+tagName,
			"Error from: "+err.Error(),
		)
		return
	}
	currentPolicies := Map(currentPoliciesInfo, func(current dtrack.TaggedPolicyListResponseItem) uuid.UUID { return current.UUID })
	desiredPolicies, err := TryMap(plan.Policies, func(value types.String) (uuid.UUID, error) {
		return uuid.Parse(value.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Update, unable to parse policy into UUID",
			"Error from: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Updating Tag Policies", map[string]any{
		"tag":     tagName,
		"current": currentPolicies,
		"desired": desiredPolicies,
	})

	addPolicies, removePolicies := ListDeltasUUID(currentPolicies, desiredPolicies)
	if len(addPolicies) > 0 {
		err = r.client.Tag.TagPolicies(ctx, tagName, addPolicies)
		if err != nil {
			resp.Diagnostics.AddError(
				"Within Update, unable to add policies from tag list for tag: "+tagName,
				"Error from: "+err.Error(),
			)
		}
	}
	if len(removePolicies) > 0 {
		err = r.client.Tag.UntagPolicies(ctx, tagName, removePolicies)
		if err != nil {
			resp.Diagnostics.AddError(
				"Within Update, unable to remove policies from tag list for tag: "+tagName,
				"Error from: "+err.Error(),
			)
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	plan = tagPoliciesResourceModel{
		ID:  types.StringValue(tagName),
		Tag: types.StringValue(tagName),
		Policies: Map(desiredPolicies, func(u uuid.UUID) types.String {
			return types.StringValue(u.String())
		}),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated Tag Policies", map[string]any{
		"tag":      plan.Tag.ValueString(),
		"policies": Map(plan.Policies, func(t types.String) string { return t.ValueString() }),
	})
}

func (r *tagPoliciesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tagPoliciesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagName := state.Tag.ValueString()
	tflog.Debug(ctx, "Deleting Tag Policies", map[string]any{
		"tag":        tagName,
		"policies.#": len(state.Policies),
	})

	currentPoliciesInfo, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.TaggedPolicyListResponseItem], error) {
		return r.client.Tag.GetPolicies(ctx, tagName, po, dtrack.SortOptions{})
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Delete, unable to request current policies for tag: "+tagName,
			"Error from: "+err.Error(),
		)
		return
	}
	currentPolicies := Map(currentPoliciesInfo, func(current dtrack.TaggedPolicyListResponseItem) uuid.UUID { return current.UUID })
	err = r.client.Tag.UntagPolicies(ctx, tagName, currentPolicies)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Delete, err removing Tag Policy, for tag: "+tagName,
			"Error from: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted Tag Policies", map[string]any{
		"tag": tagName,
		"policies": Map(state.Policies, func(policy types.String) string {
			return policy.ValueString()
		}),
	})
}

func (*tagPoliciesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing Tag Policies", map[string]any{
		"id": req.ID,
	})
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported Tag Policies", map[string]any{
		"id": req.ID,
	})
}

func (r *tagPoliciesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clientInfoData, ok := req.ProviderData.(clientInfo)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Configure Type",
			fmt.Sprintf("Expected provider.clientInfo, got %T. Please report this issue to the provider developer.", req.ProviderData),
		)
		return
	}
	r.client = clientInfoData.client
	r.semver = clientInfoData.semver
}
