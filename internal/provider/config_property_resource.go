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
	_ resource.Resource                = &configPropertyResource{}
	_ resource.ResourceWithConfigure   = &configPropertyResource{}
	_ resource.ResourceWithImportState = &configPropertyResource{}
)

func NewConfigPropertyResource() resource.Resource {
	return &configPropertyResource{}
}

type configPropertyResource struct {
	client *dtrack.Client
}

type configPropertyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Group       types.String `tfsdk:"group"`
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
}

func (r *configPropertyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_property"
}

func (r *configPropertyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Config Property.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID used by provider. Has no meaning to DependencyTrack.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group": schema.StringAttribute{
				Description: "Group name of the Config Property.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Property name of the Config Property.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Description: "Value of the Config Property.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the Config Property. See DependencyTrack for valid enum values.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the Config Property.",
				Computed:    true,
			},
		},
	}
}

func (r *configPropertyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan configPropertyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	propertyReq := dtrack.ConfigProperty{
		GroupName: plan.Group.ValueString(),
		Name:      plan.Name.ValueString(),
		Value:     plan.Value.ValueString(),
		Type:      plan.Type.ValueString(),
	}

	tflog.Debug(ctx, "Configuring a config property", map[string]any{
		"group": propertyReq.GroupName,
		"name":  propertyReq.Name,
		"value": propertyReq.Value,
		"type":  propertyReq.Type,
	})
	propertyRes, err := r.client.Config.Update(ctx, propertyReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error configuring config property.",
			"Unexpected error: "+err.Error(),
		)
		return
	}
	propertyState := configPropertyResourceModel{
		ID:          types.StringValue(fmt.Sprintf("%s/%s", propertyRes.GroupName, propertyRes.Name)),
		Group:       types.StringValue(propertyRes.GroupName),
		Name:        types.StringValue(propertyRes.Name),
		Value:       types.StringValue(propertyRes.Value),
		Type:        types.StringValue(propertyRes.Type),
		Description: types.StringValue(propertyRes.Description),
	}

	diags = resp.State.Set(ctx, &propertyState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created a managed config property")
}

func (r *configPropertyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state configPropertyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupName := state.Group.ValueString()
	propertyName := state.Name.ValueString()

	tflog.Debug(ctx, "Refreshing config property", map[string]any{
		"group": groupName,
		"name":  propertyName,
	})
	configProperty, err := r.client.Config.Get(ctx, groupName, propertyName)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("config"),
			"Within Read, unable to fetch config property.",
			"Error from: "+err.Error(),
		)
		return
	}
	propertyState := configPropertyResourceModel{
		ID:          types.StringValue(fmt.Sprintf("%s/%s", configProperty.GroupName, configProperty.Name)),
		Group:       types.StringValue(configProperty.GroupName),
		Name:        types.StringValue(configProperty.Name),
		Value:       types.StringValue(configProperty.Value),
		Type:        types.StringValue(configProperty.Type),
		Description: types.StringValue(configProperty.Description),
	}
	diags = resp.State.Set(ctx, &propertyState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Refreshed config property.")
}

func (r *configPropertyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan configPropertyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	propertyReq := dtrack.ConfigProperty{
		GroupName: plan.Group.ValueString(),
		Name:      plan.Name.ValueString(),
		Value:     plan.Value.ValueString(),
		Type:      plan.Type.ValueString(),
	}

	tflog.Debug(ctx, "Updating config property", map[string]any{
		"group": propertyReq.GroupName,
		"name":  propertyReq.Name,
		"value": propertyReq.Value,
		"type":  propertyReq.Type,
	})

	propertyRes, err := r.client.Config.Update(ctx, propertyReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update config property.",
			"Error from: "+err.Error(),
		)
		return
	}
	state := configPropertyResourceModel{
		ID:          types.StringValue(fmt.Sprintf("%s/%s", propertyRes.GroupName, propertyRes.Name)),
		Group:       types.StringValue(propertyRes.GroupName),
		Name:        types.StringValue(propertyRes.Name),
		Value:       types.StringValue(propertyRes.Value),
		Type:        types.StringValue(propertyRes.Type),
		Description: types.StringValue(propertyRes.Description),
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated config property")
}

func (r *configPropertyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state configPropertyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	propertyReq := dtrack.ConfigProperty{
		GroupName: state.Group.ValueString(),
		Name:      state.Name.ValueString(),
		Type:      state.Type.ValueString(),
		Value:     "",
	}
	tflog.Debug(ctx, "Deleting config property", map[string]any{
		"group": propertyReq.GroupName,
		"name":  propertyReq.Name,
		"type":  propertyReq.Type,
	})
	_, err := r.client.Config.Update(ctx, propertyReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete config property.",
			"Error from: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted config property.")
}

func (r *configPropertyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected import id",
			fmt.Sprintf("Expected id in format <Group>/<Name>. Received %s", req.ID),
		)
		return
	}
	groupName := idParts[0]
	propertyName := idParts[1]

	property, err := r.client.Config.Get(ctx, groupName, propertyName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Import, unable to locate config property.",
			"Unexpected error from: "+err.Error(),
		)
	}
	propertyState := configPropertyResourceModel{
		ID:          types.StringValue(fmt.Sprintf("%s/%s", property.GroupName, property.Name)),
		Group:       types.StringValue(property.GroupName),
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
	tflog.Debug(ctx, "Imported a config property.")
}

func (r *configPropertyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
