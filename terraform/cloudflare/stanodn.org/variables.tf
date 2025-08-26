variable "cf_zone_id" {
  description = "Cloudflare Zone ID for stanodn.org"
  type        = string
}

variable "cf_api_token" {
  description = "Cloudflare API Token"
  type        = string
  sensitive   = true
}