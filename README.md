# ddns
Dynamic DNS for cloudflare-hosted record using the Go Standard Library

## Running

Required variables:
- `ZONE_NAME` the name of the cloudflare-managed DNS zone
- `RECORD_NAME` the name of the record
- `CLOUDFLARE_TOKEN` bearer token with the ability to write to DNS zones

Optional variables:
- `POLL_MINUTE_INTERVAL` how often in minutes to check for IP changes. Defaults to 1