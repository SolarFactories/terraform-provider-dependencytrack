package provider

import (
	"context"
	"fmt"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Interface impl check.
var (
	_ datasource.DataSource              = &componentsDataSource{}
	_ datasource.DataSourceWithConfigure = &componentsDataSource{}
)

type (
	componentsDataSource struct {
		client *dtrack.Client
		semver *Semver
	}

	componentsDataSourceModel struct {
		Project      types.String `tfsdk:"project"`
		Components   []componentResourceModel
		OnlyDirect   types.Bool `tfsdk:"only_direct"`
		OnlyOutdated types.Bool `tfsdk:"only_outdated"`
	}
)

func NewComponentsDataSource() datasource.DataSource {
	return &componentsDataSource{}
}

func (*componentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_components"
}

func (*componentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetch a config property by group and name.",
		Attributes: map[string]schema.Attribute{
			"project": schema.StringAttribute{
				Description: "UUID of the Project for which to retrieve components.",
				Required:    true,
			},
			"only_direct": schema.StringAttribute{
				Description: "Filter for only direct components of the project.",
				Optional:    true,
			},
			"only_outdated": schema.StringAttribute{
				Description: "Filter for only outdated components of the project.",
				Optional:    true,
			},
			"components": schema.ListNestedAttribute{
				Description: "Components within the project.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "UUID of the Component.",
							Computed:    true,
						},
						"author": schema.StringAttribute{
							Description: "Author of the Component.",
							Computed:    true,
						},
						"publisher": schema.StringAttribute{
							Description: "Publisher of the Component.",
							Computed:    true,
						},
						"group": schema.StringAttribute{
							Description: "Group Name of the Component.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the Component.",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "Version of the Component.",
							Computed:    true,
						},
						"classifier": schema.StringAttribute{
							Description: "Classifier of the Component.",
							Computed:    true,
						},
						"filename": schema.StringAttribute{
							Description: "Filename of the Component.",
							Computed:    true,
						},
						"extension": schema.StringAttribute{
							Description: "Filename Extension of the Component.",
							Computed:    true,
						},
						"cpe": schema.StringAttribute{
							Description: "Common Platform Enumeration of the Component. Standardised format v2.2 / v2.3 from MITRE / NIST.",
							Computed:    true,
						},
						"purl": schema.StringAttribute{
							Description: "Package URL of the Component, in standardised form.",
							Computed:    true,
						},
						"swid": schema.StringAttribute{
							Description: "SWID Tag ID. ISO/IEC 19770-2:2015.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the Component.",
							Computed:    true,
						},
						"copyright": schema.StringAttribute{
							Description: "Copyright of the Component.",
							Computed:    true,
						},
						"license": schema.StringAttribute{
							Description: "License of the Component.",
							Computed:    true,
						},
						"notes": schema.StringAttribute{
							Description: "Notes of the Component.",
							Computed:    true,
						},
						"hashes": schema.SingleNestedAttribute{
							Description: "Hashes of the Component.",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"md5": schema.StringAttribute{
									Description: "MD5 hash of the Component.",
									Computed:    true,
								},
								"sha1": schema.StringAttribute{
									Description: "SHA1 hash of the Component.",
									Computed:    true,
								},
								"sha256": schema.StringAttribute{
									Description: "SHA256 hash of the Component.",
									Computed:    true,
								},
								"sha384": schema.StringAttribute{
									Description: "SHA384 hash of the Component.",
									Computed:    true,
								},
								"sha512": schema.StringAttribute{
									Description: "SHA512 hash of the Component.",
									Computed:    true,
								},
								"sha3_256": schema.StringAttribute{
									Description: "SHA3-256 hash of the Component.",
									Computed:    true,
								},
								"sha3_384": schema.StringAttribute{
									Description: "SHA3-384 hash of the Component.",
									Computed:    true,
								},
								"sha3_512": schema.StringAttribute{
									Description: "SHA3-512 hash of the Component.",
									Computed:    true,
								},
								"blake2b_256": schema.StringAttribute{
									Description: "BLAKE2b-256 hash of the Component.",
									Computed:    true,
								},
								"blake2b_384": schema.StringAttribute{
									Description: "BLAKE2b-384 hash of the Component.",
									Computed:    true,
								},
								"blake2b_512": schema.StringAttribute{
									Description: "BLAKE2b-512 hash of the Component.",
									Computed:    true,
								},
								"blake3": schema.StringAttribute{
									Description: "BLAKE3 hash of the Component.",
									Computed:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *componentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state componentsDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, diagnostic := TryParseUUID(state.Project, LifecycleRead, path.Root("project"))
	onlyDirect := state.OnlyDirect.ValueBool()
	onlyOutdated := state.OnlyOutdated.ValueBool()
	diags.Append(diagnostic)
	if diags.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Project Components", map[string]any{
		"project":       project.String(),
		"only_direct":   onlyDirect,
		"only_outdated": onlyOutdated,
	})

	components, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.Component], error) {
		return d.client.Component.GetAll(ctx, project, po, dtrack.ComponentFilterOptions{
			OnlyOutdated: onlyOutdated,
			OnlyDirect:   onlyDirect,
		})
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Within Read, unable to fetch Components",
			"Error from: "+err.Error(),
		)
		return
	}
	state = componentsDataSourceModel{
		OnlyDirect:   types.BoolValue(onlyDirect),
		OnlyOutdated: types.BoolValue(onlyOutdated),
		Project:      types.StringValue(project.String()),
		Components:   Map(components, componentToModel),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Read Project Components", map[string]any{
		"project": project.String(),
		"count":   len(state.Components),
	})
}

func (d *componentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
