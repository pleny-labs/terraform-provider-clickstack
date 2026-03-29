# Terraform Provider for ClickStack (HyperDX)

The ClickStack provider enables infrastructure-as-code management of your [ClickStack](https://clickhouse.com/docs/clickstack) (HyperDX) observability platform. Define dashboards, alerts, and webhooks as Terraform resources to version-control your monitoring configuration alongside your infrastructure.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- A ClickStack instance with API access enabled
- An API key with appropriate permissions

## Setup

```hcl
terraform {
  required_providers {
    clickstack = {
      source  = "pleny-labs/clickstack"
      version = "~> 0.1"
    }
  }
}

provider "clickstack" {
  endpoint = "https://clickstack.example.com"  # or CLICKSTACK_ENDPOINT env var
  api_key  = var.clickstack_api_key             # or CLICKSTACK_API_KEY env var
}

variable "clickstack_api_key" {
  type      = string
  sensitive = true
}
```

## Usage Examples

### Create a webhook for alert notifications

```hcl
resource "clickstack_webhook" "slack_alerts" {
  name        = "Slack Alerts"
  service     = "slack"
  url         = var.slack_webhook_url
  description = "Send alerts to the #ops-alerts Slack channel"
}
```

### Build a dashboard with multiple tiles

```hcl
data "clickstack_sources" "all" {}

resource "clickstack_dashboard" "service_overview" {
  name = "Service Overview"
  tags = ["production", "sre"]

  tiles {
    name = "Request Rate"
    x    = 0
    y    = 0
    w    = 6
    h    = 3

    config {
      display_type = "line"
      source_id    = data.clickstack_sources.all.sources[0].id

      select {
        agg_fn           = "count"
        value_expression = "*"
      }

      group_by = ["ServiceName"]
    }
  }

  tiles {
    name = "Error Count"
    x    = 6
    y    = 0
    w    = 6
    h    = 3

    config {
      display_type = "number"
      source_id    = data.clickstack_sources.all.sources[0].id

      select {
        agg_fn           = "count"
        value_expression = "*"
        where            = "SeverityText = 'ERROR'"
        where_language   = "sql"
        alias            = "Errors"
      }
    }
  }
}
```

### Set up an alert on a dashboard tile

```hcl
resource "clickstack_alert" "high_error_rate" {
  name           = "High Error Rate"
  threshold      = 100
  interval       = "5m"
  threshold_type = "above"
  source         = "tile"
  dashboard_id   = clickstack_dashboard.service_overview.id
  tile_id        = "tile-1"
  message        = "Error rate exceeded threshold in the last 5 minutes"

  channel {
    type       = "webhook"
    webhook_id = clickstack_webhook.slack_alerts.id
  }
}
```

### Look up existing webhooks

```hcl
data "clickstack_webhooks" "all" {}

output "available_webhooks" {
  value = data.clickstack_webhooks.all.webhooks[*].name
}
```

## Resources

| Resource                   | Description                                          |
|----------------------------|------------------------------------------------------|
| `clickstack_dashboard`     | Dashboards with tiles, filters, and saved queries    |
| `clickstack_alert`         | Threshold-based alerts with webhook notifications    |
| `clickstack_webhook`       | Webhooks for alert notification delivery             |

## Data Sources

| Data Source                | Description                                          |
|----------------------------|------------------------------------------------------|
| `clickstack_sources`       | Lists all configured data sources                    |
| `clickstack_webhooks`      | Lists all configured webhooks                        |

## Import

All resources support `terraform import`:

```bash
terraform import clickstack_dashboard.example <dashboard-id>
terraform import clickstack_alert.example <alert-id>
terraform import clickstack_webhook.example <webhook-id>
```

## Authentication

The provider supports two authentication methods:

1. **Provider configuration** (shown above)
2. **Environment variables**:
   - `CLICKSTACK_ENDPOINT` - API base URL
   - `CLICKSTACK_API_KEY` - Bearer token for authentication

Environment variables are useful for CI/CD pipelines and avoiding secrets in Terraform configuration files.

## Development

```bash
make build    # Build the provider binary
make test     # Run unit tests
make testacc  # Run acceptance tests (requires API credentials)
make docs     # Regenerate documentation
make install  # Install locally for testing
```
