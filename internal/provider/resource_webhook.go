package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pleny-labs/terraform-provider-clickstack/internal/client"
)

var (
	_ resource.Resource                = &WebhookResource{}
	_ resource.ResourceWithImportState = &WebhookResource{}
)

type WebhookResource struct {
	client *client.Client
}

type WebhookResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Service     types.String `tfsdk:"service"`
	URL         types.String `tfsdk:"url"`
	Description types.String `tfsdk:"description"`
	Body        types.String `tfsdk:"body"`
	QueryParams types.Map    `tfsdk:"query_params"`
	Headers     types.Map    `tfsdk:"headers"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewWebhookResource() resource.Resource {
	return &WebhookResource{}
}

func (r *WebhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *WebhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a ClickStack webhook.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the webhook.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the webhook.",
				Required:    true,
			},
			"service": schema.StringAttribute{
				Description: "The service type (e.g., slack, generic).",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "The webhook URL.",
				Required:    true,
				Sensitive:   true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the webhook.",
				Optional:    true,
			},
			"body": schema.StringAttribute{
				Description: "Custom request body template.",
				Optional:    true,
			},
			"query_params": schema.MapAttribute{
				Description: "Query parameters to include in webhook requests.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"headers": schema.MapAttribute{
				Description: "Custom headers to include in webhook requests.",
				Optional:    true,
				ElementType: types.StringType,
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
	}
}

func (r *WebhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	r.client = c
}

func expandWebhookFromPlan(ctx context.Context, plan WebhookResourceModel, diags *diag.Diagnostics) client.Webhook {
	w := client.Webhook{
		Name:    plan.Name.ValueString(),
		Service: plan.Service.ValueString(),
		URL:     plan.URL.ValueString(),
	}
	if !plan.Description.IsNull() {
		w.Description = plan.Description.ValueString()
	}
	if !plan.Body.IsNull() {
		w.Body = plan.Body.ValueString()
	}
	if !plan.QueryParams.IsNull() {
		qp := make(map[string]string)
		diags.Append(plan.QueryParams.ElementsAs(ctx, &qp, false)...)
		w.QueryParams = qp
	}
	if !plan.Headers.IsNull() {
		h := make(map[string]string)
		diags.Append(plan.Headers.ElementsAs(ctx, &h, false)...)
		w.Headers = h
	}
	return w
}

func flattenWebhookToState(w *client.Webhook, state *WebhookResourceModel) {
	state.ID = types.StringValue(w.ID)
	state.Name = types.StringValue(w.Name)
	state.Service = types.StringValue(w.Service)
	state.URL = types.StringValue(w.URL)
	state.Description = stringOrNull(w.Description)
	state.Body = stringOrNull(w.Body)
	state.CreatedAt = stringOrNull(w.CreatedAt)
	state.UpdatedAt = stringOrNull(w.UpdatedAt)
	state.QueryParams = types.MapNull(types.StringType)
	state.Headers = types.MapNull(types.StringType)
}

func (r *WebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiWebhook := expandWebhookFromPlan(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CreateWebhook(ctx, apiWebhook)
	if err != nil {
		resp.Diagnostics.AddError("Error creating webhook", err.Error())
		return
	}

	flattenWebhookToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhook, err := r.client.GetWebhook(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading webhook", err.Error())
		return
	}

	flattenWebhookToState(webhook, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state WebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiWebhook := expandWebhookFromPlan(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UpdateWebhook(ctx, state.ID.ValueString(), apiWebhook)
	if err != nil {
		resp.Diagnostics.AddError("Error updating webhook", err.Error())
		return
	}

	flattenWebhookToState(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteWebhook(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting webhook", err.Error())
		return
	}
}

func (r *WebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	webhook, err := r.client.GetWebhook(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing webhook", err.Error())
		return
	}

	var state WebhookResourceModel
	flattenWebhookToState(webhook, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
