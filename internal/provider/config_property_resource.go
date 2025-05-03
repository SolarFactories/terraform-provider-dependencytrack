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
	semver *Semver
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
	if propertyRes.Type == PropertyTypeEncryptedString {
		propertyState.Value = plan.Value
	}

	diags = resp.State.Set(ctx, &propertyState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created a config property", map[string]any{
		"id":          propertyState.ID.ValueString(),
		"group":       propertyState.Group.ValueString(),
		"name":        propertyState.Name.ValueString(),
		"value":       propertyState.Value.ValueString(),
		"type":        propertyState.Type.ValueString(),
		"description": propertyState.Description.ValueString(),
	})
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

	tflog.Debug(ctx, "Reading config property", map[string]any{
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
	if configProperty.Type == PropertyTypeEncryptedString {
		propertyState.Value = state.Value
	}
	diags = resp.State.Set(ctx, &propertyState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read config property", map[string]any{
		"id":          propertyState.ID.ValueString(),
		"group":       propertyState.Group.ValueString(),
		"name":        propertyState.Name.ValueString(),
		"type":        propertyState.Type.ValueString(),
		"description": propertyState.Description.ValueString(),
	})
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
	if propertyRes.Type == PropertyTypeEncryptedString {
		state.Value = plan.Value
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated config property", map[string]any{
		"id":          state.ID.ValueString(),
		"group":       state.Group.ValueString(),
		"name":        state.Name.ValueString(),
		"value":       state.Value.ValueString(),
		"type":        state.Type.ValueString(),
		"description": state.Description.ValueString(),
	})
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
	propertyRes, err := r.client.Config.Update(ctx, propertyReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to delete config property.",
			"Error from: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted config property", map[string]any{
		"id":    state.ID.ValueString(),
		"group": propertyRes.GroupName,
		"name":  propertyRes.Name,
		"value": propertyRes.Value,
		"type":  propertyRes.Type,
	})
}

func (r *configPropertyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected import id",
			"Expected id in format <Group>/<Name>. Received "+req.ID,
		)
		return
	}
	groupName := idParts[0]
	propertyName := idParts[1]
	tflog.Debug(ctx, "Importing Config Property", map[string]any{
		"group": groupName,
		"name":  propertyName,
	})

	property, err := r.client.Config.Get(ctx, groupName, propertyName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Import, unable to locate config property.",
			"Unexpected error from: "+err.Error(),
		)
		return
	}
	propertyState := configPropertyResourceModel{
		ID:    types.StringValue(fmt.Sprintf("%s/%s", property.GroupName, property.Name)),
		Group: types.StringValue(property.GroupName),
		Name:  types.StringValue(property.Name),
		// If Type == "ENCRYPTEDSTRING", then Value will be placeholder text
		Value:       types.StringValue(property.Value),
		Type:        types.StringValue(property.Type),
		Description: types.StringValue(property.Description),
	}
	diags := resp.State.Set(ctx, propertyState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported config property", map[string]any{
		"id":          propertyState.ID.ValueString(),
		"group":       propertyState.Group.ValueString(),
		"name":        propertyState.Name.ValueString(),
		"value":       propertyState.Value.ValueString(),
		"type":        propertyState.Type.ValueString(),
		"description": propertyState.Description.ValueString(),
	})
}

func (r *configPropertyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
