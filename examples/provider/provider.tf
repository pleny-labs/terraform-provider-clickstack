terraform {
  required_providers {
    clickstack = {
      source = "pleny-labs/clickstack"
    }
  }
}

# Option 1: Session cookie auth (for ClickStack instances without v2 API)
provider "clickstack" {
  endpoint       = "https://clickstack.example.com:8443"
  session_cookie = var.clickstack_session_cookie
}

# Option 2: Bearer token auth (for ClickStack instances with v2 external API)
# provider "clickstack" {
#   endpoint      = "https://clickstack.example.com:8443"
#   api_key       = var.clickstack_api_key
#   api_base_path = "/api/v2"
# }

variable "clickstack_session_cookie" {
  type      = string
  sensitive = true
}
