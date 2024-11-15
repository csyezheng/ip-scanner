# ip-scanner

[![Go Version](https://img.shields.io/github/go-mod/go-version/sinspired/ip-scanner?logo=go&label=Go)](https://github.com/sinspired/ip-scanner)
[![Release Version](https://img.shields.io/github/v/release/sinspired/ip-scanner?display_name=tag&logo=github&label=Release)](https://github.com/sinspired/ip-scanner/releases/latest)
[![GitHub repo size](https://img.shields.io/github/repo-size/sinspired/ip-scanner?logo=github)
](https://github.com/sinspired/ip-scanner)
[![GitHub last commit](https://img.shields.io/github/last-commit/sinspired/ip-scanner?logo=github&label=最后提交：)](https://github.com/sinspired/ip-scanner)

[![build and release](https://github.com/sinspired/ip-scanner/actions/workflows/create_release.yml/badge.svg)](https://github.com/sinspired/ip-scanner/actions/workflows/create_release.yml)

This script is used to find the fastest IP for a given domain name.

## Use
* GoogleTranslate 

  Google has ended its Google Translate service in mainland China. In order to continue using Google's translation service, look for available IPs.

* Cloudflare

  To better use Cloudflare pages and Cloudflare workers, find the fastest IP.

## Quick start

### Google Translate

Find available IPs for Google Translate:

```
go run cmd/google_translate/main.go
```

```
go run cmd/google_translate/main.go -config ./configs/config.toml
```

Fetch the latest IP ranges of Google Translate, save to the `IPRangesFile` path in the configuration file:

```
go run cmd/fetch_ip_ranges/main.go -site GoogleTranslate
```

### Cloudflare

Find the fastest IP for Cloudflare:

```
go run cmd/cloudflare/main.go
```

```
go run cmd/cloudflare/main.go -config ./configs/config.toml
```

Fetch the latest IP ranges of Cloudflare, save to the `IPRangesFile` path in the configuration file:

```
go run cmd/fetch_ip_ranges/main.go -site Cloudflare
```

### Custom site

Find available IPs for other websites, add configuration and run:

```shell
go run cmd/ip_scanner/main.go -site <site name>
```

Options:

```
-config string
    Config file, toml format (default "./configs/config.toml")
-site string
    site: the site name configured in the configuration file
```

## Configuration

```toml
[General]
# GoogleTranslate or Cloudflare
Site = "GoogleTranslate"
# A boolean that turns on/off debug mode. true or false
Debug = false
# workers
Workers = 300
# Limit the maximum number of IPs scanned. No limit if it is less than or equal to 0.
ScannedLimit = 0
# Limit the maximum number of IPs found. No limit if it is less than or equal to 0.
FoundLimit = 10

[Ping]
# avaivable values: icmp, tcp, udp
Protocol = "icmp"
# Port for tcp and udp, icmp will ignore port
Port = 443
# Times of tests per IP
Count = 3
# Millisecond
Timeout = 500
# true: it's legal if it succeeds every time. false: it's legal if it has one succeeds
all = false

[HTTP]
# Standard HTTPS ports are 443 and 8443.
Port = 443
# Times of tests per IP
Count = 3
# Millisecond
Timeout = 2000
# true: it's legal if it succeeds every time. false: it's legal if it has one succeeds
all = false

[[Sites]]
Name = "GoogleTranslate"
# The API to fetch the IP ranges
IPRangesAPI = "https://www.gstatic.com/ipranges/goog.json"
# All IP ranges of google
IPRangesFile = "./data/all_google_translate_ip_ranges.txt"
# Customized IP ranges. If the file does not exist, will use IPRangesFile
CustomIPRangesFile = "./data/custom_google_translate_ip_ranges.txt"
# Output the available IPs found
IPOutputFile = "./data/output_google_translate_ips.txt"
# # boolean that turns on/off scanning for IPv6. true or false.
WithIPv6 = false
# URL for testing HTTPS connection
HttpsURL = "https://translate.google.com"
# Domains for write into hosts file
Domains = ["translate.google.com", "translate.googleapis.com"]

[[Sites]]
Name = "Cloudflare"
# The API to fetch the IP ranges
IPRangesAPI = "https://api.cloudflare.com/client/v4/ips"
# All IP ranges of cloudflare
IPRangesFile = "./data/all_cloudflare_ip_ranges.txt"
# Customized IP ranges. If the file does not exist, will use IPRangesFile
CustomIPRangesFile = "./data/custom_cloudflare_ip_ranges.txt"
# Output the available IPs found
IPOutputFile = "./data/output_cloudflare_ips.txt"
# A boolean that turns on/off scanning for IPv6. true or false.
WithIPv6 = false
# URL for testing HTTPS connection
HttpsURL = "https://yezheng.pages.dev"
# Domains for write into hosts file
Domains = ["yezheng.pages.dev"]
```

## IP address ranges
### [Obtain Google IP address ranges](https://support.google.com/a/answer/10026322?hl=en)
* [IP ranges that Google makes available to users on the internet](https://www.gstatic.com/ipranges/goog.json)
* [Global and regional external IP address ranges for customers' Google Cloud resources](https://www.gstatic.com/ipranges/cloud.json)

### [Cloudflare IP Ranges](https://www.cloudflare.com/ips/)
* [ips-v4](https://www.cloudflare.com/ips-v4/)
* [ips-v6](https://www.cloudflare.com/ips-v6/)
* [API](https://api.cloudflare.com/client/v4/ips)

