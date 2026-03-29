package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/pleny-labs/terraform-provider-clickstack/internal/client"
)

var (
	_ resource.Resource                = &AlertResource{}
	_ resource.ResourceWithImportState = &AlertResource{}
)

type AlertResource struct {
	client *client.Client
}

type AlertResourceModel struct {
	ID                    types.String  `tfsdk:"id"`
	Name                  types.String  `tfsdk:"name"`
	Threshold             types.Float64 `tfsdk:"threshold"`
	Interval              types.String  `tfsdk:"interval"`
	ThresholdType         types.String  `tfsdk:"threshold_type"`
	Source                types.String  `tfsdk:"source"`
	Channel               types.Object  `tfsdk:"channel"`
	DashboardID           types.String  `tfsdk:"dashboard_id"`
	TileID                types.String  `tfsdk:"tile_id"`
	SavedSearchID         types.String  `tfsdk:"saved_search_id"`
	GroupBy               types.String  `tfsdk:"group_by"`
	ScheduleOffsetMinutes types.Int64   `tfsdk:"schedule_offset_minutes"`
	ScheduleStartAt       types.String  `tfsdk:"schedule_start_at"`
	Message               types.String  `tfsdk:"message"`
	State                 types.String  `tfsdk:"state"`
	TeamID                types.String  `tfsdk:"team_id"`
	CreatedAt             types.String  `tfsdk:"created_at"`
	UpdatedAt             types.String  `tfsdk:"updated_at"`
}

type AlertChannelModel struct {
	Type      types.String `tfsdk:"type"`
	WebhookID types.String `tfsdk:"webhook_id"`
}

var alertChannelAttrTypes = map[string]attr.Type{
	"type":       types.StringType,
	"webhook_id": types.StringType,
}

func NewAlertResource() resource.Resource {
	return &AlertResource{}
}

func (r *AlertResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_alert"
}

func (r *AlertResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a ClickStack alert.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the alert.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Display name of the alert.",
				Optional:    true,
			},
			"threshold": schema.Float64Attribute{
				Description: "Trigger threshold value.",
				Required:    true,
			},
			"interval": schema.StringAttribute{
				Description: "Check interval: 1m, 5m, 15m, 30m, 1h, 6h, 12h, or 1d.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("1m", "5m", "15m", "30m", "1h", "6h", "12h", "1d"),
				},
			},
			"threshold_type": schema.StringAttribute{
				Description: "Threshold direction: above or below.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("above", "below"),
				},
			},
			"source": schema.StringAttribute{
				Description: "Alert source: saved_search or tile.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("saved_search", "tile"),
				},
			},
			"channel": schema.SingleNestedAttribute{
				Description: "Notification channel configuration.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Channel type (webhook).",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("webhook"),
						},
					},
					"webhook_id": schema.StringAttribute{
						Description: "Webhook identifier for notifications.",
						Required:    true,
					},
				},
			},
			"dashboard_id": schema.StringAttribute{
				Description: "Dashboard ID for tile-based alerts.",
				Optional:    true,
			},
			"tile_id": schema.StringAttribute{
				Description: "Tile ID for tile-based alerts.",
				Optional:    true,
			},
			"saved_search_id": schema.StringAttribute{
				Description: "Saved search ID for saved_search-based alerts.",
				Optional:    true,
			},
			"group_by": schema.StringAttribute{
				Description: "Field to group alert results by.",
				Optional:    true,
			},
			"schedule_offset_minutes": schema.Int64Attribute{
				Description: "Schedule offset in minutes.",
				Optional:    true,
			},
			"schedule_start_at": schema.StringAttribute{
				Description: "Schedule start time (ISO 8601).",
				Optional:    true,
			},
			"message": schema.StringAttribute{
				Description: "Alert message template.",
				Optional:    true,
			},
			"state": schema.StringAttribute{
				Description: "Current alert state: ALERT, OK, INSUFFICIENT_DATA, or DISABLED.",
				Computed:    true,
			},
			"team_id": schema.StringAttribute{
				Description: "Owner team identifier.",
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
	}
}

func (r *AlertResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func expandAlertChannel(ctx context.Context, obj types.Object, diags *diag.Diagnostics) client.AlertChannel {
	var ch AlertChannelModel
	diags.Append(obj.As(ctx, &ch, basetypes.ObjectAsOptions{})...)
	return client.AlertChannel{
		Type:      ch.Type.ValueString(),
		WebhookID: ch.WebhookID.ValueString(),
	}
}

func flattenAlertChannel(ch client.AlertChannel) types.Object {
	obj, _ := types.ObjectValue(alertChannelAttrTypes, map[string]attr.Value{
		"type":       types.StringValue(ch.Type),
		"webhook_id": types.StringValue(ch.WebhookID),
	})
	return obj
}

func expandAlert(ctx context.Context, plan AlertResourceModel, diags *diag.Diagnostics) client.Alert {
	channel := expandAlertChannel(ctx, plan.Channel, diags)

	a := client.Alert{
		Threshold:     plan.Threshold.ValueFloat64(),
		Interval:      plan.Interval.ValueString(),
		ThresholdType: plan.ThresholdType.ValueString(),
		Source:        plan.Source.ValueString(),
		Channel:       channel,
	}

	if !plan.Name.IsNull() {
		v := plan.Name.ValueString()
		a.Name = &v
	}
	if !plan.DashboardID.IsNull() {
		v := plan.DashboardID.ValueString()
		a.DashboardID = &v
	}
	if !plan.TileID.IsNull() {
		v := plan.TileID.ValueString()
		a.TileID = &v
	}
	if !plan.SavedSearchID.IsNull() {
		v := plan.SavedSearchID.ValueString()
		a.SavedSearchID = &v
	}
	if !plan.GroupBy.IsNull() {
		v := plan.GroupBy.ValueString()
		a.GroupBy = &v
	}
	if !plan.ScheduleOffsetMinutes.IsNull() {
		v := int(plan.ScheduleOffsetMinutes.ValueInt64())
		a.ScheduleOffsetMinutes = &v
	}
	if !plan.ScheduleStartAt.IsNull() {
		v := plan.ScheduleStartAt.ValueString()
		a.ScheduleStartAt = &v
	}
	if !plan.Message.IsNull() {
		v := plan.Message.ValueString()
		a.Message = &v
	}

	return a
}

func flattenAlert(alert *client.Alert, state *AlertResourceModel) {
	state.ID = types.StringValue(alert.ID)
	state.Threshold = types.Float64Value(alert.Threshold)
	state.Interval = types.StringValue(alert.Interval)
	state.ThresholdType = types.StringValue(alert.ThresholdType)
	state.Source = types.StringValue(alert.Source)
	state.Channel = flattenAlertChannel(alert.Channel)
	state.State = types.StringValue(alert.State)
	state.TeamID = types.StringValue(alert.TeamID)
	state.CreatedAt = types.StringValue(alert.CreatedAt)
	state.UpdatedAt = types.StringValue(alert.UpdatedAt)

	if alert.Name != nil {
		state.Name = types.StringValue(*alert.Name)
	} else {
		state.Name = types.StringNull()
	}
	if alert.DashboardID != nil {
		state.DashboardID = types.StringValue(*alert.DashboardID)
	} else {
		state.DashboardID = types.StringNull()
	}
	if alert.TileID != nil {
		state.TileID = types.StringValue(*alert.TileID)
	} else {
		state.TileID = types.StringNull()
	}
	if alert.SavedSearchID != nil {
		state.SavedSearchID = types.StringValue(*alert.SavedSearchID)
	} else {
		state.SavedSearchID = types.StringNull()
	}
	if alert.GroupBy != nil {
		state.GroupBy = types.StringValue(*alert.GroupBy)
	} else {
		state.GroupBy = types.StringNull()
	}
	if alert.ScheduleOffsetMinutes != nil {
		state.ScheduleOffsetMinutes = types.Int64Value(int64(*alert.ScheduleOffsetMinutes))
	} else {
		state.ScheduleOffsetMinutes = types.Int64Null()
	}
	if alert.ScheduleStartAt != nil {
		state.ScheduleStartAt = types.StringValue(*alert.ScheduleStartAt)
	} else {
		state.ScheduleStartAt = types.StringNull()
	}
	if alert.Message != nil {
		state.Message = types.StringValue(*alert.Message)
	} else {
		state.Message = types.StringNull()
	}
}

func (r *AlertResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiAlert := expandAlert(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.CreateAlert(ctx, apiAlert)
	if err != nil {
		resp.Diagnostics.AddError("Error creating alert", err.Error())
		return
	}

	flattenAlert(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AlertResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	alert, err := r.client.GetAlert(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading alert", err.Error())
		return
	}

	flattenAlert(alert, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AlertResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AlertResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiAlert := expandAlert(ctx, plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.UpdateAlert(ctx, state.ID.ValueString(), apiAlert)
	if err != nil {
		resp.Diagnostics.AddError("Error updating alert", err.Error())
		return
	}

	flattenAlert(result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AlertResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AlertResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAlert(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting alert", err.Error())
		return
	}
}

func (r *AlertResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	alert, err := r.client.GetAlert(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing alert", err.Error())
		return
	}

	var state AlertResourceModel
	flattenAlert(alert, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
