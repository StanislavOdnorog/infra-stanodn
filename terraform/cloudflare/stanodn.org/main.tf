# Read public DNS records from JSON file
locals {
  dns_records = jsondecode(file("${path.module}/dns-records.json"))
}

# Create regular DNS records (A, CNAME, MX, TXT)
resource "cloudflare_dns_record" "dns_records" {
  for_each = {
    for idx, record in local.dns_records : "${record.name}-${record.type}" => record
    if record.type != "SRV"
  }

  zone_id  = var.cf_zone_id
  name     = each.value.name
  type     = each.value.type
  content  = each.value.value
  ttl      = lookup(each.value, "proxied", false) ? 1 : lookup(each.value, "ttl", 1)
  priority = lookup(each.value, "priority", null)
  proxied  = lookup(each.value, "proxied", false)
  comment  = lookup(each.value, "comment", "Terraform managed")
}

# Create SRV records with data block
resource "cloudflare_dns_record" "srv_records" {
  for_each = {
    for idx, record in local.dns_records : "${record.name}-${record.type}" => record
    if record.type == "SRV"
  }

  zone_id = var.cf_zone_id
  name    = each.value.name
  type    = "SRV"
  ttl     = 1
  comment = lookup(each.value, "comment", "Terraform managed")

  data = {
    priority = tonumber(split(" ", each.value.value)[0])
    weight   = tonumber(split(" ", each.value.value)[1])
    port     = tonumber(split(" ", each.value.value)[2])
    target   = split(" ", each.value.value)[3]
  }
}