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
	_ resource.Resource                = &tagProjectsResource{}
	_ resource.ResourceWithConfigure   = &tagProjectsResource{}
	_ resource.ResourceWithImportState = &tagProjectsResource{}
)

type (
	tagProjectsResource struct {
		client *dtrack.Client
		semver *Semver
	}

	tagProjectsResourceModel struct {
		ID       types.String   `tfsdk:"id"`
		Tag      types.String   `tfsdk:"tag"`
		Projects []types.String `tfsdk:"projects"`
	}
)

func NewTagProjectsResource() resource.Resource {
	return &tagProjectsResource{}
}

func (*tagProjectsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag_projects"
}

func (*tagProjectsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Applies an existing tag to multiple projects. Requires API version >= 4.12.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Name of the Tag.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tag": schema.StringAttribute{
				Description: "Name of the Tag. Must be lowercase.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"projects": schema.ListAttribute{
				Description: "Project UUIDs to which to apply tag. Sorted by project name.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *tagProjectsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tagProjectsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagName := plan.Tag.ValueString()
	currentProjectsInfo, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.TaggedProjectListResponseItem], error) {
		return r.client.Tag.GetProjects(ctx, tagName, po, dtrack.SortOptions{})
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Create, unable to request current projects for tag: "+tagName,
			"Error from: "+err.Error(),
		)
		return
	}
	currentProjects := Map(currentProjectsInfo, func(current dtrack.TaggedProjectListResponseItem) uuid.UUID { return current.UUID })
	desiredProjects, err := TryMap(plan.Projects, func(value types.String) (uuid.UUID, error) {
		return uuid.Parse(value.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Create, unable to parse project into UUID",
			"Error from: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Creating Tag Projects", map[string]any{
		"tag":     tagName,
		"current": currentProjects,
		"desired": desiredProjects,
	})

	addProjects, removeProjects := ListDeltasUUID(currentProjects, desiredProjects)
	if len(addProjects) > 0 {
		err = r.client.Tag.TagProjects(ctx, tagName, addProjects)
		if err != nil {
			resp.Diagnostics.AddError(
				"Within Create, unable to add projects from tag list for tag: "+tagName,
				"Error from: "+err.Error(),
			)
		}
	}
	if len(removeProjects) > 0 {
		err = r.client.Tag.UntagProjects(ctx, tagName, removeProjects)
		if err != nil {
			resp.Diagnostics.AddError(
				"Within Create, unable to remove projects from tag list for tag: "+tagName,
				"Error from: "+err.Error(),
			)
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	plan = tagProjectsResourceModel{
		ID:  types.StringValue(tagName),
		Tag: types.StringValue(tagName),
		Projects: Map(desiredProjects, func(project uuid.UUID) types.String {
			return types.StringValue(project.String())
		}),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created Tag Projects", map[string]any{
		"id":       plan.ID.ValueString(),
		"tag":      plan.Tag.ValueString(),
		"projects": Map(plan.Projects, func(item types.String) string { return item.ValueString() }),
	})
}

func (r *tagProjectsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tagProjectsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagName := state.ID.ValueString()
	tflog.Debug(ctx, "Reading Tag Projects", map[string]any{
		"id":         state.ID.ValueString(),
		"tag":        tagName,
		"projects.#": len(state.Projects),
		"projects":   Map(state.Projects, func(item types.String) string { return item.ValueString() }),
	})

	taggedProjectsInfo, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.TaggedProjectListResponseItem], error) {
		return r.client.Tag.GetProjects(ctx, tagName, po, dtrack.SortOptions{})
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to get current list of projects for tag: "+tagName,
			"Error from: "+err.Error(),
		)
		return
	}

	state = tagProjectsResourceModel{
		ID:  types.StringValue(tagName),
		Tag: types.StringValue(tagName),
		Projects: Map(taggedProjectsInfo, func(info dtrack.TaggedProjectListResponseItem) types.String {
			return types.StringValue(info.UUID.String())
		}),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Tag Projects", map[string]any{
		"id":       state.ID.ValueString(),
		"tag":      state.Tag.ValueString(),
		"projects": Map(state.Projects, func(v types.String) string { return v.ValueString() }),
	})
}

func (r *tagProjectsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tagProjectsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagName := plan.Tag.ValueString()
	tflog.Debug(ctx, "Updating Tag Projects", map[string]any{
		"id":         plan.ID.ValueString(),
		"tag":        tagName,
		"projects.#": len(plan.Projects),
		"projects":   Map(plan.Projects, func(item types.String) string { return item.ValueString() }),
	})

	currentProjectsInfo, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.TaggedProjectListResponseItem], error) {
		return r.client.Tag.GetProjects(ctx, tagName, po, dtrack.SortOptions{})
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Update, unable to request current projects for tag: "+tagName,
			"Error from: "+err.Error(),
		)
		return
	}
	currentProjects := Map(currentProjectsInfo, func(current dtrack.TaggedProjectListResponseItem) uuid.UUID { return current.UUID })
	desiredProjects, err := TryMap(plan.Projects, func(value types.String) (uuid.UUID, error) {
		return uuid.Parse(value.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Update, unable to parse project into UUID",
			"Error from: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Updating Tag Projects", map[string]any{
		"id":      plan.ID.ValueString(),
		"tag":     tagName,
		"current": currentProjects,
		"desired": desiredProjects,
	})

	addProjects, removeProjects := ListDeltasUUID(currentProjects, desiredProjects)
	if len(addProjects) > 0 {
		err = r.client.Tag.TagProjects(ctx, tagName, addProjects)
		if err != nil {
			resp.Diagnostics.AddError(
				"Within Update, unable to add projects from tag list for tag: "+tagName,
				"Error from: "+err.Error(),
			)
		}
	}
	if len(removeProjects) > 0 {
		err = r.client.Tag.UntagProjects(ctx, tagName, removeProjects)
		if err != nil {
			resp.Diagnostics.AddError(
				"Within Update, unable to remove projects from tag list for tag: "+tagName,
				"Error from: "+err.Error(),
			)
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	plan = tagProjectsResourceModel{
		ID:  types.StringValue(tagName),
		Tag: types.StringValue(tagName),
		Projects: Map(desiredProjects, func(u uuid.UUID) types.String {
			return types.StringValue(u.String())
		}),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated Tag Projects", map[string]any{
		"id":       plan.ID.ValueString(),
		"tag":      plan.Tag.ValueString(),
		"projects": Map(plan.Projects, func(t types.String) string { return t.ValueString() }),
	})
}

func (r *tagProjectsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tagProjectsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagName := state.Tag.ValueString()
	tflog.Debug(ctx, "Deleting Tag Projects", map[string]any{
		"id":         state.ID.ValueString(),
		"tag":        tagName,
		"projects.#": len(state.Projects),
		"projects":   Map(state.Projects, func(t types.String) string { return t.ValueString() }),
	})

	currentProjectsInfo, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.TaggedProjectListResponseItem], error) {
		return r.client.Tag.GetProjects(ctx, tagName, po, dtrack.SortOptions{})
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Delete, unable to request current projects for tag: "+tagName,
			"Error from: "+err.Error(),
		)
		return
	}
	currentProjects := Map(currentProjectsInfo, func(current dtrack.TaggedProjectListResponseItem) uuid.UUID { return current.UUID })
	if len(currentProjects) > 0 {
		err = r.client.Tag.UntagProjects(ctx, tagName, currentProjects)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Delete, err removing Tag Projects, for tag: "+tagName,
			"Error from: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted Tag Projects", map[string]any{
		"id":  state.ID.ValueString(),
		"tag": tagName,
		"projects": Map(state.Projects, func(project types.String) string {
			return project.ValueString()
		}),
	})
}

func (*tagProjectsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing Tag Projects", map[string]any{
		"id": req.ID,
	})
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported Tag Projects", map[string]any{
		"id": req.ID,
	})
}

func (r *tagProjectsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
