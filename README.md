# walrss

*Email-based RSS digest generator*

[![RSS is cool](https://img.shields.io/badge/rss-is%20cool-orange?logo=rss)](https://github.com/codemicro/walrss/releases.atom) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/codemicro/walrss) [![Go Report Card](https://goreportcard.com/badge/github.com/codemicro/walrss)](https://goreportcard.com/report/github.com/codemicro/walrss) ![Lines of code](https://img.shields.io/tokei/lines/github/codemicro/walrss)

---

| Email                                                | Web UI                                   |
|------------------------------------------------------|------------------------------------------|
| ![Sample email](.github/screenshots/sampleEmail.png) | ![Web UI](.github/screenshots/webUI.png) |

## Features

* Multi-user support
* Option to choose between daily or weekly digests that are sent at specific times of day
* OPML import/export
* [OIDC](https://en.wikipedia.org/wiki/OpenID#OpenID_Connect_(OIDC)) integration

## Quickstart

### Docker

```bash
docker pull ghcr.io/codemicro/walrss:latest
mkdir walrss-data
nano walrss-data/config.yaml  # You'll have to fill in the required fields detailed below
docker run -d -v $(pwd)/walrss-data:/run -p 8080:8080 ghcr.io/codemicro/walrss:latest
```

### Build from source

You must have an up-to-date version of the Go toolchain installed. Check `go.mod` for the minimum required version.

```bash
git clone https://github.com/codemicro/walrss.git; cd walrss
go build -o walrssServer github.com/codemicro/walrss/walrss
```

```bash
mkdir walrss-data
nano walrss-data/config.yaml  # You'll have to fill in the required fields detailed below
WALRSS_DIR="./walrss-data" ./walrssServer
# now go to http://127.0.0.1:8080
```

## Config

Your `config.yaml` file lives in your data directory, which is specified by the `WALRSS_DIR` environment variable (default is `./` or `/run` if you're using Docker). Below is a list of the possible options you can set within it.

```yaml
server:
  host: 127.0.0.1
  port: 8080
  # externalURL is the URL that your instance of Walrss is accessible at
  # This is a required field
  externalURL: "http://127.0.0.1:8080"
email:
  # These are required fields
  host: "smtp.sendgrid.net"
  port: 587
  from: "Walrss <walrss@yourdomain.com>"
  # These fields are not required
  tls: "" # valid values are: starttls, tls, no - default: starttls
  password: "yourapikey"
  username: "apikey"
platform:
  disableRegistration: false
  disableSecureCookies: false
  contactInformation: "https://example.com"  # optional. will be included in the user agent if set.
oidc:
  enable: false
  clientID: "yourclientid"
  clientSecret: "yourclientsecret"
  issuer: "https://sso.example.com/"
```

## Future features

* Feed categories
* Administrator users
* Account controls (eg password reset, change password, etc)
