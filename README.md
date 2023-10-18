# ip-scanner

This script is used to find the fastest IP for a given domain name.

## Use
* Cloudflare

  To better use cloudflare pages and cloudflare workers, find the fastest IP.

* GoogleTransle 

  To reduce the impact of these disturbances, find clean IPs (IPs that are not disturbed).


## Run
```shell
go run .\cmd\main.go
```
Options:
```
-config string
    Config file, toml format (default "./configs/config.toml")
-domain string
    domain: cloudflare or google (default "cloudflare")
```

## Configuration
```toml
[General]
Debug = true
# workers
Workers = 100
Domain = "cloudflare"

[Ping]
# avaivable values: icmp, tcp, udp
Protocol = "icmp"
# Port for tcp and udp, icmp will ignore port
Port = 443
# Times of tests per IP
Count = 3
# Millisecond
Timeout = 500

[HTTP]
# Standard HTTPS ports are 443 and 8443.
Port = 443
# Millisecond
Timeout = 2000

[Domains]

[Domains.Cloudflare]
IPRangesFile = "./data/cloudflare.json"
IPOutputFile = "./data/cloudflare.txt"
WithIPv6 = false
HttpsURL = "https://yezheng.pages.dev"

[Domains.GoogleTranslate]
IPRangesFile = "./data/goog.json"
IPOutputFile = "./data/goog.txt"
WithIPv6 = false
HttpsURL = "https://translate.google.com"
```

## IP address ranges
### [Obtain Google IP address ranges](https://support.google.com/a/answer/10026322?hl=en)
* [IP ranges that Google makes available to users on the internet](https://www.gstatic.com/ipranges/goog.json)
* [Global and regional external IP address ranges for customers' Google Cloud resources](https://www.gstatic.com/ipranges/cloud.json)

### [Cloudflare IP Ranges](https://www.cloudflare.com/ips/)
* [ips-v4](https://www.cloudflare.com/ips-v4/)
* [ips-v6](https://www.cloudflare.com/ips-v6/)
* [API](https://api.cloudflare.com/client/v4/ips)
