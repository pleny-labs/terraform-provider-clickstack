package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pleny-labs/terraform-provider-clickstack/internal/client"
)

var _ datasource.DataSource = &WebhooksDataSource{}

type WebhooksDataSource struct {
	client *client.Client
}

type WebhooksDataSourceModel struct {
	Webhooks types.List `tfsdk:"webhooks"`
}

var webhookAttrTypes = map[string]attr.Type{
	"id":          types.StringType,
	"name":        types.StringType,
	"service":     types.StringType,
	"url":         types.StringType,
	"description": types.StringType,
	"created_at":  types.StringType,
	"updated_at":  types.StringType,
}

func NewWebhooksDataSource() datasource.DataSource {
	return &WebhooksDataSource{}
}

func (d *WebhooksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhooks"
}

func (d *WebhooksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all ClickStack webhooks.",
		Attributes: map[string]schema.Attribute{
			"webhooks": schema.ListNestedAttribute{
				Description: "List of webhooks.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Webhook identifier.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Webhook name.",
							Computed:    true,
						},
						"service": schema.StringAttribute{
							Description: "Service type.",
							Computed:    true,
						},
						"url": schema.StringAttribute{
							Description: "Webhook URL.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Webhook description.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Creation timestamp.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "Last update timestamp.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *WebhooksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *WebhooksDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	webhooks, err := d.client.ListWebhooks(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading webhooks", err.Error())
		return
	}

	webhookValues := make([]attr.Value, len(webhooks))
	for i, w := range webhooks {
		obj, diags := types.ObjectValue(webhookAttrTypes, map[string]attr.Value{
			"id":          types.StringValue(w.ID),
			"name":        types.StringValue(w.Name),
			"service":     types.StringValue(w.Service),
			"url":         types.StringValue(w.URL),
			"description": types.StringValue(w.Description),
			"created_at":  types.StringValue(w.CreatedAt),
			"updated_at":  types.StringValue(w.UpdatedAt),
		})
		resp.Diagnostics.Append(diags...)
		webhookValues[i] = obj
	}

	webhooksList, diags := types.ListValue(types.ObjectType{AttrTypes: webhookAttrTypes}, webhookValues)
	resp.Diagnostics.Append(diags...)

	state := WebhooksDataSourceModel{Webhooks: webhooksList}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
