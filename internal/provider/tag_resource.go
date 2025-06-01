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
	_ resource.Resource                = &tagResource{}
	_ resource.ResourceWithConfigure   = &tagResource{}
	_ resource.ResourceWithImportState = &tagResource{}
)

type (
	tagResource struct {
		client *dtrack.Client
		semver *Semver
	}

	tagResourceModel struct {
		ID   types.String `tfsdk:"id"`
		Name types.String `tfsdk:"name"`
	}
)

func NewTagResource() resource.Resource {
	return &tagResource{}
}

func (*tagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (*tagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Tag. Requires API version >= 4.13 to be created, but may be imported in earlier versions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Name of the Tag.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Tag.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *tagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tagResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tagName := plan.Name.ValueString()

	tflog.Debug(ctx, "Creating Tag", map[string]any{
		"name": tagName,
	})
	err := r.client.Tag.Create(ctx, []string{tagName})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Tag",
			"Error from: "+err.Error(),
		)
		return
	}
	plan = tagResourceModel{
		ID:   types.StringValue(tagName),
		Name: types.StringValue(tagName),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created Tag", map[string]any{
		"id":   plan.ID.ValueString(),
		"name": plan.Name.ValueString(),
	})
}

func (r *tagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state.
	var state tagResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refresh.
	tagID := state.ID.ValueString()
	tflog.Debug(ctx, "Reading Tag", map[string]any{
		"id":   state.ID.ValueString(),
		"name": state.Name.ValueString(),
	})

	tag, err := FindPaged(func(po dtrack.PageOptions) (dtrack.Page[dtrack.TagListResponseItem], error) {
		return r.client.Tag.GetAll(ctx, po, dtrack.SortOptions{})
	}, func(tag dtrack.TagListResponseItem) bool {
		return tag.Name == tagID
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get updated tag",
			"Error with reading tag: "+tagID+", from: "+err.Error(),
		)
		return
	}
	state = tagResourceModel{
		ID:   types.StringValue(tag.Name),
		Name: types.StringValue(tag.Name),
	}

	// Update state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Tag", map[string]any{
		"id":   state.ID.ValueString(),
		"name": state.Name.ValueString(),
	})
}

func (*tagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Resource has no Update action, as any changes require replacement
	// Get State.
	var plan tagResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK.
	tagID := plan.ID.ValueString()
	// Execute.
	tflog.Debug(ctx, "Updating Tag", map[string]any{
		"id":   tagID,
		"name": tagID,
	})
	// Map SDK to TF.
	plan = tagResourceModel{
		ID:   types.StringValue(tagID),
		Name: types.StringValue(tagID),
	}

	// Update State.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated Tag", map[string]any{
		"id":   plan.ID.ValueString(),
		"name": plan.Name.ValueString(),
	})
}

func (r *tagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state.
	var state tagResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK.
	tagID := state.ID.ValueString()

	// Execute.
	tflog.Debug(ctx, "Deleting Tag", map[string]any{
		"id":   state.ID.ValueString(),
		"name": state.Name.ValueString(),
	})
	err := r.client.Tag.Delete(ctx, []string{tagID})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete tag",
			"Unexpected error when trying to delete tag: "+tagID+", from error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted Tag", map[string]any{
		"id":   state.ID.ValueString(),
		"name": state.Name.ValueString(),
	})
}

func (*tagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing Tag", map[string]any{
		"id": req.ID,
	})
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported Tag", map[string]any{
		"id": req.ID,
	})
}

func (r *tagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
