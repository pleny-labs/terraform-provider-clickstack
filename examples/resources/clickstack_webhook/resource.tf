resource "clickstack_webhook" "slack_alerts" {
  name        = "Slack Alerts"
  service     = "slack"
  url         = var.slack_webhook_url
  description = "Send alerts to the #ops-alerts Slack channel"
}

variable "slack_webhook_url" {
  type      = string
  sensitive = true
}
