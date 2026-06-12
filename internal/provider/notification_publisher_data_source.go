package provider

import (
	"context"
	"fmt"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Interface impl check.
var (
	_ datasource.DataSource              = &notificationPublisherDataSource{}
	_ datasource.DataSourceWithConfigure = &notificationPublisherDataSource{}
)

type (
	notificationPublisherDataSource struct {
		client *dtrack.Client
		semver *Semver
	}

	notificationPublisherDataSourceModel struct {
		ID               types.String `tfsdk:"id"`
		Name             types.String `tfsdk:"name"`
		Description      types.String `tfsdk:"description"`
		PublisherClass   types.String `tfsdk:"publisher_class"`
		Template         types.String `tfsdk:"template"`
		TemplateMimeType types.String `tfsdk:"template_mime_type"`
		DefaultPublisher types.Bool   `tfsdk:"default_publisher"`
	}
)

func NewNotificationPublisherDataSource() datasource.DataSource {
	return &notificationPublisherDataSource{}
}

func (*notificationPublisherDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_publisher"
}

func (*notificationPublisherDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetch an existing Notification Publisher by Name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "UUID of the Notification Publisher located.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the Notification Publisher to find.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Publisher Description.",
				Computed:    true,
			},
			"publisher_class": schema.StringAttribute{
				Description: "Name of Java Class that provides Publisher.",
				Computed:    true,
			},
			"template": schema.StringAttribute{
				Description: "Template string value for Publisher Payload.",
				Computed:    true,
			},
			"template_mime_type": schema.StringAttribute{
				Description: "MIME type set when sending a notification, for template.",
				Computed:    true,
			},
			"default_publisher": schema.BoolAttribute{
				Description: "Whether this is a default publisher.",
				Computed:    true,
			},
		},
	}
}

func (d *notificationPublisherDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state notificationPublisherDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Reading Notification Publisher", map[string]any{
		"name": state.Name.ValueString(),
	})

	publishers, err := d.client.Notification.GetAllPublishers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Notification Publishers",
			"Error when fetching publishers: "+err.Error(),
		)
		return
	}

	publisher, err := Find(publishers, func(pub dtrack.NotificationPublisher) bool {
		return pub.Name == state.Name.ValueString()
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to locate Notification Publisher",
			"Error with locating Name: "+state.Name.ValueString()+", in original error: "+err.Error(),
		)
		return
	}

	// Transform data into model.
	state = notificationPublisherDataSourceModel{
		ID:               types.StringValue(publisher.UUID.String()),
		Name:             types.StringValue(publisher.Name),
		Description:      types.StringValue(publisher.Description),
		PublisherClass:   types.StringValue(publisher.PublisherClass),
		Template:         types.StringValue(publisher.Template),
		TemplateMimeType: types.StringValue(publisher.TemplateMIMEType),
		DefaultPublisher: types.BoolValue(publisher.DefaultPublisher),
	}

	// Update state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Notification Publisher", map[string]any{
		"id":   state.ID.ValueString(),
		"name": state.Name.ValueString(),
	})
}

func (d *notificationPublisherDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = clientInfoData.client
	d.semver = clientInfoData.semver
}
