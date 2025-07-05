package provider

import (
	"context"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &componentResource{}
	_ resource.ResourceWithConfigure   = &componentResource{}
	_ resource.ResourceWithImportState = &componentResource{}
)

type (
	componentResource struct {
		client *dtrack.Client
		semver *Semver
	}

	componentResourceModel struct {
		ID          types.String                 `tfsdk:"id"`
		Author      types.String                 `tfsdk:"author"`
		Publisher   types.String                 `tfsdk:"publisher"`
		Group       types.String                 `tfsdk:"group"`
		Name        types.String                 `tfsdk:"name"`
		Version     types.String                 `tfsdk:"version"`
		Classifier  types.String                 `tfsdk:"classifier"`
		Filename    types.String                 `tfsdk:"filename"`
		Extension   types.String                 `tfsdk:"extension"`
		CPE         types.String                 `tfsdk:"cpe"`
		PURL        types.String                 `tfsdk:"purl"`
		SWID        types.String                 `tfsdk:"swid"`
		Description types.String                 `tfsdk:"description"`
		Copyright   types.String                 `tfsdk:"copyright"`
		License     types.String                 `tfsdk:"license"`
		Notes       types.String                 `tfsdk:"notes"`
		Hashes      componentHashesResourceModel `tfsdk:"hashes"`
		Internal    types.Bool                   `tfsdk:"internal"`
	}

	componentHashesResourceModel struct {
		MD5         types.String `tfsdk:"md5"`
		SHA1        types.String `tfsdk:"sha1"`
		SHA256      types.String `tfsdk:"sha256"`
		SHA384      types.String `tfsdk:"sha384"`
		SHA512      types.String `tfsdk:"sha512"`
		SHA3_256    types.String `tfsdk:"sha3_256"`
		SHA3_384    types.String `tfsdk:"sha3_384"`
		SHA3_512    types.String `tfsdk:"sha_512"`
		BLAKE2b_256 types.String `tfsdk:"blake2b_256"`
		BLAKE2b_284 types.String `tfsdk:"blake2b_384"`
		BLAKE2b_512 types.String `tfsdk:"bake2b_512"`
		BLAKE3      types.String `tfsdk:"blake3"`
	}
)

func NewComponentResource() resource.Resource {
	return &componentResource{}
}

func (*componentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_component"
}

func (*componentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// TODO
}

func (r *componentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// TODO
}

func (r *componentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// TODO
}

func (r *componentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO
}

func (r *componentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// TODO
}

func (r *componentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// TODO
}

func (r *componentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// TODO
}
