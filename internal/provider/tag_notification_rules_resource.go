package provider

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource                = &tagNotificationRulesResource{}
	_ resource.ResourceWithConfigure   = &tagNotificationRulesResource{}
	_ resource.ResourceWithImportState = &tagNotificationRulesResource{}
)

type (
	tagNotificationRulesResource struct {
		client *dtrack.Client
		semver *Semver
	}

	tagNotificationRulesResourceModel struct {
		ID                types.String   `tfsdk:"id"`
		Tag               types.String   `tfsdk:"tag"`
		NotificationRules []types.String `tfsdk:"notification_rules"`
	}
)

func NewTagNotificationRulesResource() resource.Resource {
	return &tagNotificationRulesResource{}
}

func (*tagNotificationRulesResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag_notification_rules"
}

func (*tagNotificationRulesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Applies an existing tag to multiple notification rules. Requires API version >= 4.12.",
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
			"notification_rules": schema.ListAttribute{
				Description: "Notification Rule UUIDs to which to apply tag.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *tagNotificationRulesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tagNotificationRulesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagName := plan.Tag.ValueString()
	currentNotificationRulesInfo, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.TaggedPolicyListResponseItem], error) {
		return r.client.Tag.GetNotificationRules(ctx, tagName, po, dtrack.SortOptions{})
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Create, unable to request current notification rules for tag: "+tagName,
			"Error from: "+err.Error(),
		)
		return
	}
	currentNotificationRules := Map(currentNotificationRulesInfo, func(current dtrack.TaggedPolicyListResponseItem) uuid.UUID { return current.UUID })
	desiredNotificationRules, err := TryMap(plan.NotificationRules, func(value types.String) (uuid.UUID, error) {
		return uuid.Parse(value.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Create, unable to parse notification rule into UUID",
			"Error from: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Creating Tag NotificationRules", map[string]any{
		"tag":     tagName,
		"current": currentNotificationRules,
		"desired": desiredNotificationRules,
	})

	addNotificationRules, removeNotificationRules := ListDeltasUUID(currentNotificationRules, desiredNotificationRules)
	if len(addNotificationRules) > 0 {
		err = r.client.Tag.TagNotificationRules(ctx, tagName, addNotificationRules)
		if err != nil {
			resp.Diagnostics.AddError(
				"Within Create, unable to add notification rules from tag list for tag: "+tagName,
				"Error from: "+err.Error(),
			)
		}
	}
	if len(removeNotificationRules) > 0 {
		err = r.client.Tag.UntagNotificationRules(ctx, tagName, removeNotificationRules)
		if err != nil {
			resp.Diagnostics.AddError(
				"Within Create, unable to remove notification rules from tag list for tag: "+tagName,
				"Error from: "+err.Error(),
			)
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	plan = tagNotificationRulesResourceModel{
		ID:  types.StringValue(tagName),
		Tag: types.StringValue(tagName),
		NotificationRules: Map(desiredNotificationRules, func(rule uuid.UUID) types.String {
			return types.StringValue(rule.String())
		}),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Created Tag NotificationRules", map[string]any{
		"id":                 plan.ID.ValueString(),
		"tag":                plan.Tag.ValueString(),
		"notification_rules": Map(plan.NotificationRules, types.String.ValueString),
	})
}

func (r *tagNotificationRulesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tagNotificationRulesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagName := state.ID.ValueString()
	tflog.Debug(ctx, "Reading Tag NotificationRules", map[string]any{
		"tag":                  tagName,
		"notification_rules.#": len(state.NotificationRules),
		"notification_rules":   Map(state.NotificationRules, types.String.ValueString),
	})

	taggedNotificationRulesInfo, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.TaggedPolicyListResponseItem], error) {
		return r.client.Tag.GetNotificationRules(ctx, tagName, po, dtrack.SortOptions{})
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to get current list of notification rules for tag: "+tagName,
			"Error from: "+err.Error(),
		)
		return
	}

	stateNotificationRules := Map(taggedNotificationRulesInfo, func(info dtrack.TaggedPolicyListResponseItem) types.String {
		return types.StringValue(info.UUID.String())
	})

	if SliceUnorderedEqual(stateNotificationRules, state.NotificationRules, func(a, b types.String) int {
		return strings.Compare(a.ValueString(), b.ValueString())
	}) {
		stateNotificationRules = state.NotificationRules
	}

	newState := tagNotificationRulesResourceModel{
		ID:                types.StringValue(tagName),
		Tag:               types.StringValue(tagName),
		NotificationRules: stateNotificationRules,
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Tag NotificationRules", map[string]any{
		"id":                 newState.ID.ValueString(),
		"tag":                newState.Tag.ValueString(),
		"notification_rules": Map(newState.NotificationRules, types.String.ValueString),
	})
}

func (r *tagNotificationRulesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tagNotificationRulesResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagName := plan.Tag.ValueString()
	tflog.Debug(ctx, "Updating Tag NotificationRules", map[string]any{
		"id":                   plan.ID.ValueString(),
		"tag":                  tagName,
		"notification_rules.#": len(plan.NotificationRules),
		"notification_rules":   Map(plan.NotificationRules, types.String.ValueString),
	})

	currentNotificationRulesInfo, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.TaggedPolicyListResponseItem], error) {
		return r.client.Tag.GetNotificationRules(ctx, tagName, po, dtrack.SortOptions{})
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Update, unable to request current notification_rules for tag: "+tagName,
			"Error from: "+err.Error(),
		)
		return
	}
	currentNotificationRules := Map(currentNotificationRulesInfo, func(current dtrack.TaggedPolicyListResponseItem) uuid.UUID { return current.UUID })
	desiredNotificationRules, err := TryMap(plan.NotificationRules, func(value types.String) (uuid.UUID, error) {
		return uuid.Parse(value.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Update, unable to parse element in notification_rules into UUID",
			"Error from: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Updating Tag NotificationRules", map[string]any{
		"id":      plan.ID.ValueString(),
		"tag":     tagName,
		"current": currentNotificationRules,
		"desired": desiredNotificationRules,
	})

	addNotificationRules, removeNotificationRules := ListDeltasUUID(currentNotificationRules, desiredNotificationRules)
	if len(addNotificationRules) > 0 {
		err = r.client.Tag.TagNotificationRules(ctx, tagName, addNotificationRules)
		if err != nil {
			resp.Diagnostics.AddError(
				"Within Update, unable to add notification rules from tag list for tag: "+tagName,
				"Error from: "+err.Error(),
			)
		}
	}
	if len(removeNotificationRules) > 0 {
		err = r.client.Tag.UntagNotificationRules(ctx, tagName, removeNotificationRules)
		if err != nil {
			resp.Diagnostics.AddError(
				"Within Update, unable to remove notification rules from tag list for tag: "+tagName,
				"Error from: "+err.Error(),
			)
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	plan = tagNotificationRulesResourceModel{
		ID:  types.StringValue(tagName),
		Tag: types.StringValue(tagName),
		NotificationRules: Map(desiredNotificationRules, func(u uuid.UUID) types.String {
			return types.StringValue(u.String())
		}),
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Updated Tag NotificationRules", map[string]any{
		"id":                 plan.ID.ValueString(),
		"tag":                plan.Tag.ValueString(),
		"notification_rules": Map(plan.NotificationRules, types.String.ValueString),
	})
}

func (r *tagNotificationRulesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tagNotificationRulesResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tagName := state.Tag.ValueString()
	tflog.Debug(ctx, "Deleting Tag NotificationRules", map[string]any{
		"id":                   state.ID.ValueString(),
		"tag":                  tagName,
		"notification_rules.#": len(state.NotificationRules),
		"notification_rules":   Map(state.NotificationRules, types.String.ValueString),
	})

	currentNotificationRulesInfo, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.TaggedPolicyListResponseItem], error) {
		return r.client.Tag.GetNotificationRules(ctx, tagName, po, dtrack.SortOptions{})
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Delete, unable to request current notification rules for tag: "+tagName,
			"Error from: "+err.Error(),
		)
		return
	}
	currentNotificationRules := Map(currentNotificationRulesInfo, func(current dtrack.TaggedPolicyListResponseItem) uuid.UUID { return current.UUID })
	if len(currentNotificationRules) > 0 {
		err = r.client.Tag.UntagNotificationRules(ctx, tagName, currentNotificationRules)
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Delete, err removing Tag Notification Rules, for tag: "+tagName,
			"Error from: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted Tag NotificationRules", map[string]any{
		"id":                 state.ID.ValueString(),
		"tag":                tagName,
		"notification_rules": Map(state.NotificationRules, types.String.ValueString),
	})
}

func (*tagNotificationRulesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, "Importing Tag NotificationRules", map[string]any{
		"id": req.ID,
	})
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Imported Tag NotificationRules", map[string]any{
		"id": req.ID,
	})
}

func (r *tagNotificationRulesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
