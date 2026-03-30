# Terraform Provider for ClickStack (HyperDX)

The ClickStack provider enables infrastructure-as-code management of your [ClickStack](https://clickhouse.com/docs/clickstack) (HyperDX) observability platform. Define dashboards, alerts, and webhooks as Terraform resources to version-control your monitoring configuration alongside your infrastructure.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- A ClickStack instance with API access

## Setup

```hcl
terraform {
  required_providers {
    clickstack = {
      source  = "pleny-labs/clickstack"
      version = ">= 0.1.18"
    }
  }
}
```

## Authentication

The provider supports two authentication methods depending on your ClickStack version:

### Session Cookie (internal API)

For ClickStack instances without the v2 external API, authenticate using a session cookie from your browser:

```hcl
provider "clickstack" {
  endpoint       = "https://clickstack.example.com:8443"
  session_cookie = var.clickstack_session_cookie  # or CLICKSTACK_SESSION_COOKIE env var
}
```

Get the cookie value from your browser's developer tools (Application > Cookies > `connect.sid`).

### Bearer Token (v2 external API)

For ClickStack instances with the v2 external API enabled:

```hcl
provider "clickstack" {
  endpoint      = "https://clickstack.example.com:8443"
  api_key       = var.clickstack_api_key       # or CLICKSTACK_API_KEY env var
  api_base_path = "/api/v2"                     # or CLICKSTACK_API_BASE_PATH env var
}
```

## Usage Examples

### Look up data sources

```hcl
data "clickstack_sources" "all" {}
```

### Create a webhook for alert notifications

```hcl
resource "clickstack_webhook" "slack_alerts" {
  name        = "Slack Alerts"
  service     = "slack"
  url         = var.slack_webhook_url
  description = "Send alerts to the #ops-alerts Slack channel"
}
```

### Build a dashboard with tiles

```hcl
resource "clickstack_dashboard" "service_overview" {
  name = "Service Overview"
  tags = ["production", "sre"]

  tiles {
    x = 0
    y = 0
    w = 12
    h = 8

    config {
      name           = "Request Rate"
      display_type   = "line"
      source         = data.clickstack_sources.all.sources[0].id
      group_by       = "ServiceName"
      where_language = "lucene"
      granularity    = "5 minute"

      select {
        agg_fn           = "count"
        value_expression = ""
      }
    }
  }

  tiles {
    x = 12
    y = 0
    w = 12
    h = 8

    config {
      name           = "Error Count"
      display_type   = "number"
      source         = data.clickstack_sources.all.sources[0].id
      where_language = "lucene"

      select {
        agg_fn                 = "count"
        value_expression       = ""
        agg_condition          = "SeverityText:ERROR"
        agg_condition_language = "lucene"
      }
    }
  }
}
```

### Set up an alert

```hcl
resource "clickstack_alert" "high_error_rate" {
  name           = "High Error Rate"
  threshold      = 100
  interval       = "5m"
  threshold_type = "above"
  source         = "tile"
  dashboard_id   = clickstack_dashboard.service_overview.id
  tile_id        = "tile-id"
  message        = "Error rate exceeded threshold"

  channel {
    type       = "webhook"
    webhook_id = clickstack_webhook.slack_alerts.id
  }
}
```

## Resources

| Resource | Description |
|---|---|
| `clickstack_dashboard` | Dashboards with tiles (line, table, number, search, markdown) |
| `clickstack_alert` | Threshold-based alerts with webhook notifications |
| `clickstack_webhook` | Webhooks for alert notification delivery (slack, generic, incidentio) |

## Data Sources

| Data Source | Description |
|---|---|
| `clickstack_sources` | Lists all configured data sources (log, trace, metric, session) |
| `clickstack_webhooks` | Lists all configured webhooks |

## Tile Configuration

Tiles use a 24-column grid layout. Each tile has position (`x`, `y`) and size (`w`, `h`).

The `config` block supports:

| Field | Description |
|---|---|
| `name` | Tile display name |
| `display_type` | `line`, `table`, `number`, `search`, or `markdown` |
| `source` | Data source ID |
| `group_by` | Field to group by (string, e.g. `"ServiceName"`) |
| `where` / `where_language` | Tile-level filter (lucene or sql) |
| `granularity` | Time bucket size (e.g. `"5 minute"`, `"1 hour"`) |
| `sort_order` | `asc` or `desc` (for table tiles) |

The `select` block within config defines aggregations:

| Field | Description |
|---|---|
| `agg_fn` | `count`, `avg`, `sum`, `min`, `max`, `quantile`, `count_distinct`, `last_value` |
| `value_expression` | Column to aggregate (empty string for count) |
| `agg_condition` | Filter for this aggregation (e.g. `"SeverityText:ERROR"`) |
| `agg_condition_language` | `lucene` or `sql` |
| `level` | Percentile for quantile (e.g. `0.95`) |

## Import

All resources support `terraform import`:

```bash
terraform import clickstack_dashboard.example <dashboard-id>
terraform import clickstack_alert.example <alert-id>
terraform import clickstack_webhook.example <webhook-id>
```

## Development

```bash
make build    # Build the provider binary
make test     # Run unit tests
make testacc  # Run acceptance tests (requires API credentials)
make docs     # Regenerate documentation
make install  # Install locally for testing
```
