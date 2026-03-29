package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pleny-labs/terraform-provider-clickstack/internal/client"
)

var _ provider.Provider = &ClickStackProvider{}

type ClickStackProvider struct {
	version string
}

type ClickStackProviderModel struct {
	Endpoint    types.String `tfsdk:"endpoint"`
	APIKey      types.String `tfsdk:"api_key"`
	APIBasePath types.String `tfsdk:"api_base_path"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ClickStackProvider{
			version: version,
		}
	}
}

func (p *ClickStackProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "clickstack"
	resp.Version = p.version
}

func (p *ClickStackProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for ClickStack (HyperDX) observability platform.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Description: "The ClickStack API base URL. Can also be set via the CLICKSTACK_ENDPOINT environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The ClickStack API key for authentication. Can also be set via the CLICKSTACK_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_base_path": schema.StringAttribute{
				Description: "The API base path prefix for all requests (e.g. /api or /api/v2). Defaults to /api. Can also be set via the CLICKSTACK_API_BASE_PATH environment variable.",
				Optional:    true,
			},
		},
	}
}

func (p *ClickStackProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ClickStackProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := os.Getenv("CLICKSTACK_ENDPOINT")
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}
	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing API Endpoint",
			"The provider cannot create the ClickStack API client because the endpoint is missing. "+
				"Set the endpoint in the provider configuration or via the CLICKSTACK_ENDPOINT environment variable.",
		)
		return
	}

	apiKey := os.Getenv("CLICKSTACK_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The provider cannot create the ClickStack API client because the API key is missing. "+
				"Set the api_key in the provider configuration or via the CLICKSTACK_API_KEY environment variable.",
		)
		return
	}

	apiBasePath := os.Getenv("CLICKSTACK_API_BASE_PATH")
	if !config.APIBasePath.IsNull() {
		apiBasePath = config.APIBasePath.ValueString()
	}

	c := client.NewClient(endpoint, apiKey, apiBasePath)
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *ClickStackProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDashboardResource,
		NewAlertResource,
		NewWebhookResource,
	}
}

func (p *ClickStackProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSourcesDataSource,
		NewWebhooksDataSource,
	}
}
