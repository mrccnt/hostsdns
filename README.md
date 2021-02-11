![License:MIT](https://img.shields.io/static/v1?label=License&message=MIT&color=green)

# HostsDNS

A very simple DNS Proxy.

Proxy through (forward) all types of requests/queries we can not handle ourselves and answer A and AAAA queries for
hosts we know. 

## Background

Hosting your own DNS server in your networks to simplify communication between your internal hosts is something we
sooner or later stumble into when dealing with lots of networks and setups. Sure there are lots of nice possibilities
out there like CoreDNS, Bind9, Pihole including filtering etc., but they need to be configured accordingly with networks,
domains, zones, and so on. I just do not want to waste time just because I want to make my little `kodi` rasperry pi
available in my home network. I also do not want to think about traffic a real DNS server can produce. I always have the
need of a service where I can say "Here are some hosts and IPs - please serve these hosts as given.". So here we go: HostsDNS

## Configuration

```json
{
  "bind": "0.0.0.0",
  "port": 53,
  "dns": "1.1.1.1:53",
  "records": {
    "myhost1": "192.168.1.61",
    "myhost2": "192.168.1.62"
  }
}
```

HostsDNS will respond with known IP from `records` for A and AAAA queries or forward queries to given `dns` server.
That's it. HostsDNS will not try to sync with other nameservers or handle any other background actions.

## Build

```shell
CGO_ENABLED=0 go build -i -a -installsuffix cgo -o dist/hostsdns *.go
docker build -t marcocontiorg/hostsdns .
```

## Testing around

Let's Assume `hostsdns` is running on 192.168.1.194

We know "myhost1" so hostsdns will respond by itself:

```shell
nslookup myhost1 192.168.1.100
#> Server:         192.168.1.194
#> Address:        192.168.1.194#53
#> 
#> Non-authoritative answer:
#> Name:   myhost1
#> Address: 192.168.1.61
```

We do not know anything about "google.com", so the query/request will be forwarded to cloudflare DNS:

```shell
nslookup google.com 192.168.1.100
#> Server:         192.168.1.194
#> Address:        192.168.1.194#53
#> 
#> Non-authoritative answer:
#> Name:   google.com
#> Address: 216.58.215.238
#> Name:   google.com
#> Address: 2a00:1450:400a:802::200e
```
