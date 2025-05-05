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
	_ resource.Resource                = &projectPropertyResource{}
	_ resource.ResourceWithConfigure   = &projectPropertyResource{}
	_ resource.ResourceWithImportState = &projectPropertyResource{}
)

func NewProjectPropertyResource() resource.Resource {
	return &projectPropertyResource{}
}

type projectPropertyResource struct {
	client *dtrack.Client
	semver *Semver
}

type projectPropertyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Project     types.String `tfsdk:"project"`
	Group       types.String `tfsdk:"group"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
}

func (r *projectPropertyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_property"
}

func (r *projectPropertyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Project Property.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID used by provider. Has no meaning to DependencyTrack.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project": schema.StringAttribute{
				Description: "UUID for the Project in which to create the Property.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group": schema.StringAttribute{
				Description: "Group name of the Project Property.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Property name of the Project Property.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Description: "Value of the Project Property.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the Project Property. See DependencyTrack for valid enum values.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the Project Property.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *projectPropertyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectPropertyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, diag := TryParseUUID(plan.Project, LifecycleCreate, path.Root("project"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	propertyReq := dtrack.ProjectProperty{
		Group:       plan.Group.ValueString(),
		Name:        plan.Name.ValueString(),
		Value:       plan.Value.ValueString(),
		Type:        plan.Type.ValueString(),
		Description: plan.Description.ValueString(),
	}

	tflog.Debug(ctx, "Creating a Project Property", map[string]any{
		"project": project.String(),
		"group":   propertyReq.Group,
		"name":    propertyReq.Name,
		"value":   propertyReq.Value,
		"type":    propertyReq.Type,
	})
	propertyRes, err := r.client.ProjectProperty.Create(ctx, project, propertyReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Project Property.",
			"Error from: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", project.String(), propertyRes.Group, propertyRes.Name))
	plan.Project = types.StringValue(project.String())
	plan.Group = types.StringValue(propertyRes.Group)
	plan.Name = types.StringValue(propertyRes.Name)
	if propertyReq.Type != PropertyTypeEncryptedString {
		plan.Value = types.StringValue(propertyRes.Value)
	}
	plan.Type = types.StringValue(propertyRes.Type)
	plan.Description = types.StringValue(propertyRes.Description)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created a Project Property", map[string]any{
		"id":          plan.ID.ValueString(),
		"project":     plan.Project.ValueString(),
		"group":       plan.Group.ValueString(),
		"name":        plan.Name.ValueString(),
		"value":       plan.Value.ValueString(),
		"type":        plan.Type.ValueString(),
		"description": plan.Description.ValueString(),
	})
}

func (r *projectPropertyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectPropertyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, diag := TryParseUUID(state.Project, LifecycleRead, path.Root("project"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	tflog.Debug(ctx, "Reading Project Property", map[string]any{
		"id":          state.ID.ValueString(),
		"project":     project.String(),
		"group":       state.Group.ValueString(),
		"name":        state.Group.ValueString(),
		"value":       state.Value.ValueString(),
		"type":        state.Type.ValueString(),
		"description": state.Description.ValueString(),
	})
	property, err := FindPaged(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.ProjectProperty], error) {
			return r.client.ProjectProperty.GetAll(ctx, project, po)
		},
		func(property dtrack.ProjectProperty) bool {
			if property.Group != state.Group.ValueString() {
				return false
			}
			if property.Name != state.Name.ValueString() {
				return false
			}
			return true
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to locate Project Property.",
			"Error from: "+err.Error(),
		)
	}
	propertyState := projectPropertyResourceModel{
		ID:          types.StringValue(fmt.Sprintf("%s/%s/%s", project.String(), property.Group, property.Name)),
		Project:     types.StringValue(project.String()),
		Group:       types.StringValue(property.Group),
		Name:        types.StringValue(property.Name),
		Value:       types.StringValue(property.Value),
		Type:        types.StringValue(property.Type),
		Description: types.StringValue(property.Description),
	}
	if property.Type == PropertyTypeEncryptedString {
		propertyState.Value = state.Value
	}
	diags = resp.State.Set(ctx, &propertyState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Project Property", map[string]any{
		"id":          propertyState.ID.ValueString(),
		"project":     propertyState.Project.ValueString(),
		"group":       propertyState.Group.ValueString(),
		"name":        propertyState.Name.ValueString(),
		"value":       propertyState.Value.ValueString(),
		"type":        propertyState.Type.ValueString(),
		"description": propertyState.Description.ValueString(),
	})
}

func (r *projectPropertyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectPropertyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	project, diag := TryParseUUID(plan.Project, LifecycleUpdate, path.Root("project"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	propertyReq := dtrack.ProjectProperty{
		Group:       plan.Group.ValueString(),
		Name:        plan.Name.ValueString(),
		Value:       plan.Value.ValueString(),
		Type:        plan.Type.ValueString(),
		Description: plan.Description.ValueString(),
	}

	tflog.Debug(ctx, "Updating Project Property", map[string]any{
		"id":          plan.ID.ValueString(),
		"project":     project.String(),
		"group":       propertyReq.Group,
		"name":        propertyReq.Name,
		"value":       propertyReq.Value,
		"type":        propertyReq.Type,
		"description": propertyReq.Description,
	})
	propertyRes, err := r.client.ProjectProperty.Update(ctx, project, propertyReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update Project Property.",
			"Error from: "+err.Error(),
		)
		return
	}
	state := projectPropertyResourceModel{
		ID:          types.StringValue(fmt.Sprintf("%s/%s/%s", project.String(), propertyRes.Group, propertyRes.Name)),
		Project:     types.StringValue(project.String()),
		Group:       types.StringValue(propertyRes.Group),
		Name:        types.StringValue(propertyRes.Name),
		Value:       types.StringValue(propertyRes.Value),
		Type:        types.StringValue(propertyRes.Type),
		Description: types.StringValue(propertyRes.Description),
	}
	if propertyRes.Type == PropertyTypeEncryptedString {
		state.Value = plan.Value
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated Project Property", map[string]any{
		"id":          state.ID.ValueString(),
		"project":     state.Project.ValueString(),
		"group":       state.Group.ValueString(),
		"name":        state.Name.ValueString(),
		"value":       state.Value.ValueString(),
		"type":        state.Type.ValueString(),
		"description": state.Description.ValueString(),
	})
}

func (r *projectPropertyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectPropertyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	project, diag := TryParseUUID(state.Project, LifecycleDelete, path.Root("project"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	groupName := state.Group.ValueString()
	propertyName := state.Name.ValueString()

	tflog.Debug(ctx, "Deleting Project Property", map[string]any{
		"id":          state.ID.ValueString(),
		"project":     project.String(),
		"group":       groupName,
		"name":        propertyName,
		"value":       state.Value.ValueString(),
		"type":        state.Type.ValueString(),
		"description": state.Description.ValueString(),
	})
	// NOTE: Has a patch applied in `http_client.go`.
	err := r.client.ProjectProperty.Delete(ctx, project, groupName, propertyName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete project property.",
			"Error from: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted Project Property", map[string]any{
		"id":          state.ID.ValueString(),
		"project":     state.Project.ValueString(),
		"group":       state.Group.ValueString(),
		"name":        state.Name.ValueString(),
		"value":       state.Value.ValueString(),
		"type":        state.Type.ValueString(),
		"description": state.Description.ValueString(),
	})
}

func (r *projectPropertyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected import id",
			"Expected id in format <Project UUID>/<Group>/<Name>. Received "+req.ID,
		)
		return
	}
	project, diag := TryParseUUID(types.StringValue(idParts[0]), LifecycleImport, path.Root("id"))
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	groupName := idParts[1]
	propertyName := idParts[2]
	tflog.Debug(ctx, "Importing Project Property", map[string]any{
		"id":      req.ID,
		"project": project.String(),
		"group":   groupName,
		"name":    propertyName,
	})

	property, err := FindPaged(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.ProjectProperty], error) {
			return r.client.ProjectProperty.GetAll(ctx, project, po)
		},
		func(property dtrack.ProjectProperty) bool {
			if property.Group != groupName {
				return false
			}
			if property.Name != propertyName {
				return false
			}
			return true
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Import, unable to locate project property.",
			"Unexpected error from: "+err.Error(),
		)
	}
	propertyState := projectPropertyResourceModel{
		ID:          types.StringValue(fmt.Sprintf("%s/%s/%s", project.String(), property.Group, property.Name)),
		Project:     types.StringValue(project.String()),
		Group:       types.StringValue(property.Group),
		Name:        types.StringValue(property.Name),
		Value:       types.StringValue(property.Value),
		Type:        types.StringValue(property.Type),
		Description: types.StringValue(property.Description),
	}
	diags := resp.State.Set(ctx, propertyState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported Project Property", map[string]any{
		"id":          propertyState.ID.ValueString(),
		"project":     propertyState.Project.ValueString(),
		"group":       propertyState.Group.ValueString(),
		"name":        propertyState.Name.ValueString(),
		"value":       propertyState.Value.ValueString(),
		"type":        propertyState.Type.ValueString(),
		"description": propertyState.Description.ValueString(),
	})
}

func (r *projectPropertyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
