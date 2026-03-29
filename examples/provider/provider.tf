terraform {
  required_providers {
    clickstack = {
      source = "pleny-labs/clickstack"
    }
  }
}

provider "clickstack" {
  endpoint = "https://clickstack.example.com"
  api_key  = var.clickstack_api_key
}

variable "clickstack_api_key" {
  type      = string
  sensitive = true
}
