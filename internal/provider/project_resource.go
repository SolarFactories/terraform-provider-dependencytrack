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
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

type projectResource struct {
	client *dtrack.Client
}

type projectResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Active      types.Bool   `tfsdk:"active"`
	Version     types.String `tfsdk:"version"`
	Parent      types.String `tfsdk:"parent"`
	Classifier  types.String `tfsdk:"classifier"`
	Group       types.String `tfsdk:"group"`
	PURL        types.String `tfsdk:"purl"`
	CPE         types.String `tfsdk:"cpe"`
	SWID        types.String `tfsdk:"swid"`
}

func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID for the Project as generated by DependencyTrack.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Project.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the Project.",
				Optional:    true,
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the Project is active. Defaults to true.",
				Optional:    true,
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "Version of the project.",
				Optional:    true,
				Computed:    true,
			},
			"parent": schema.StringAttribute{
				Description: "UUID of a parent project, to allow for nesting. Available in API 4.12+.",
				Optional:    true,
			},
			"classifier": schema.StringAttribute{
				Description: "Classifier of the Project. Defaults to APPLICATION. See DependencyTrack for valid options.",
				Optional:    true,
				Computed:    true,
			},
			"group": schema.StringAttribute{
				Description: "Namespace / group / vendor of the Project.",
				Optional:    true,
				Computed:    true,
			},
			"purl": schema.StringAttribute{
				Description: "Package URL of the Project. MUST be in standardised format to be saved. See DependencyTrack for format.",
				Optional:    true,
				Computed:    true,
			},
			"cpe": schema.StringAttribute{
				Description: "Common Platform Enumeration of the Project. Standardised format v2.2 / v2.3 from MITRE / NIST.",
				Optional:    true,
				Computed:    true,
			},
			"swid": schema.StringAttribute{
				Description: "SWID Tag ID. ISO/IEC 19770-2:2015.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectReq := dtrack.Project{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Active:      plan.Active.ValueBool(),
		Version:     plan.Version.ValueString(),
		ParentRef:   nil,
		Classifier:  plan.Classifier.ValueString(),
		Group:       plan.Group.ValueString(),
		PURL:        plan.PURL.ValueString(),
		CPE:         plan.CPE.ValueString(),
		SWIDTagID:   plan.SWID.ValueString(),
	}
	if !plan.Parent.IsNull() {
		parentID, err := uuid.Parse(plan.Parent.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("parent"),
				"Could not parse parent id into UUID.",
				"Error from: "+err.Error(),
			)
			return
		}
		projectReq.ParentRef = &dtrack.ParentRef{
			UUID: parentID,
		}
	}
	if plan.Active.IsUnknown() {
		projectReq.Active = true
	}
	if plan.Classifier.IsUnknown() {
		projectReq.Classifier = "APPLICATION"
	}

	tflog.Debug(ctx, "Creating a new project, with name: "+projectReq.Name)
	projectRes, err := r.client.Project.Create(ctx, projectReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project",
			"Could not create project, unexpected error within Client: "+err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(projectRes.UUID.String())
	plan.Name = types.StringValue(projectRes.Name)
	plan.Description = types.StringValue(projectRes.Description)
	plan.Active = types.BoolValue(projectRes.Active)
	plan.Version = types.StringValue(projectRes.Version)
	plan.Classifier = types.StringValue(projectRes.Classifier)
	plan.Group = types.StringValue(projectRes.Group)
	plan.PURL = types.StringValue(projectRes.PURL)
	plan.CPE = types.StringValue(projectRes.CPE)
	plan.SWID = types.StringValue(projectRes.SWIDTagID)
	if projectRes.ParentRef != nil {
		plan.Parent = types.StringValue(projectRes.ParentRef.UUID.String())
	} else {
		plan.Parent = types.StringNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created a new project, with id: "+projectRes.UUID.String())
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Fetch state
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refresh
	tflog.Debug(ctx, "Refreshing project with id: "+state.ID.ValueString())
	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Within Read, unable to parse id into UUID",
			"Error from: "+err.Error(),
		)
		return
	}
	project, err := r.client.Project.Get(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get updated project",
			"Error with reading project: "+id.String()+", in original error: "+err.Error(),
		)
		return
	}
	state.ID = types.StringValue(project.UUID.String())
	state.Name = types.StringValue(project.Name)
	state.Description = types.StringValue(project.Description)
	state.Active = types.BoolValue(project.Active)
	state.Version = types.StringValue(project.Version)
	state.Classifier = types.StringValue(project.Classifier)
	state.Group = types.StringValue(project.Group)
	state.PURL = types.StringValue(project.PURL)
	state.CPE = types.StringValue(project.CPE)
	state.SWID = types.StringValue(project.SWIDTagID)
	if project.ParentRef != nil {
		state.Parent = types.StringValue(project.ParentRef.UUID.String())
	} else {
		state.Parent = types.StringNull()
	}

	// Update state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Refreshed project with id: "+state.ID.ValueString())
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get State
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK
	id, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Within Update, unable to parse id into UUID",
			"Error from: "+err.Error(),
		)
		return
	}
	project, err := r.client.Project.Get(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Update, unable to retrieve current Project",
			"Error from: "+err.Error(),
		)
		return
	}
	project.Name = plan.Name.ValueString()
	project.Description = plan.Description.ValueString()
	project.Active = plan.Active.ValueBool()
	project.Version = plan.Version.ValueString()
	project.Classifier = plan.Classifier.ValueString()
	project.Group = plan.Group.ValueString()
	project.PURL = plan.PURL.ValueString()
	project.CPE = plan.CPE.ValueString()
	project.SWIDTagID = plan.SWID.ValueString()

	if plan.Active.IsUnknown() {
		project.Active = true
	}
	if plan.Classifier.IsUnknown() {
		project.Classifier = "APPLICATION"
	}
	if !plan.Parent.IsNull() {
		parentID, err := uuid.Parse(plan.Parent.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				path.Root("parent"),
				"Unable to parse parent ID into UUID",
				"Error from: "+err.Error(),
			)
			return
		}
		project.ParentRef = &dtrack.ParentRef{
			UUID: parentID,
		}
	} else {
		project.ParentRef = nil
	}

	// Execute
	tflog.Debug(ctx, "Updating project with id: "+id.String())
	projectRes, err := r.client.Project.Update(ctx, project)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update project",
			"Error in: "+id.String()+", from: "+err.Error(),
		)
		return
	}

	// Map SDK to TF
	plan.ID = types.StringValue(projectRes.UUID.String())
	plan.Name = types.StringValue(projectRes.Name)
	plan.Description = types.StringValue(projectRes.Description)
	plan.Active = types.BoolValue(projectRes.Active)
	plan.Version = types.StringValue(projectRes.Version)
	plan.Classifier = types.StringValue(projectRes.Classifier)
	plan.Group = types.StringValue(projectRes.Group)
	plan.PURL = types.StringValue(projectRes.PURL)
	plan.CPE = types.StringValue(projectRes.CPE)
	plan.SWID = types.StringValue(projectRes.SWIDTagID)
	if projectRes.ParentRef != nil {
		plan.Parent = types.StringValue(projectRes.ParentRef.UUID.String())
	} else {
		plan.Parent = types.StringNull()
	}

	// Update State
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated project with id: "+id.String())
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Load state
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Map TF to SDK
	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Within Delete, unable to parse UUID",
			"Error parsing UUID from: "+state.ID.ValueString()+", error: "+err.Error(),
		)
		return
	}

	// Execute
	tflog.Debug(ctx, "Deleting project with id: "+id.String())
	err = r.client.Project.Delete(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete project",
			"Unexpected error when trying to delete project: "+id.String()+", error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted project with id: "+id.String())
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*dtrack.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Configure Type",
			fmt.Sprintf("Expected *dtrack.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = client
}
