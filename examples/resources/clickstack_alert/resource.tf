resource "clickstack_alert" "high_error_rate" {
  name           = "High Error Rate"
  threshold      = 100
  interval       = "5m"
  threshold_type = "above"
  source         = "tile"
  dashboard_id   = clickstack_dashboard.example.id
  tile_id        = "tile-1"
  message        = "Error rate exceeded threshold"

  channel {
    type       = "webhook"
    webhook_id = clickstack_webhook.slack_alerts.id
  }
}
