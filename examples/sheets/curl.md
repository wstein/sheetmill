---
title: curl
description: Transfer data to and from URLs from the command line, with a focus on HTTP requests, uploads and scripting.
category: networking
tags: [curl, http, networking, api, cli]
cheatsheet:
  tool_version: "8.7.1"
  updated: "2026-10-14"
source:
  license: EUPL-1.2
  attribution: SheetMill contributors
  note: Independently authored from the curl 8.7.1 documentation; examples are displayed, never executed.
---
curl sends one request per URL and writes the response body to standard output unless told otherwise. The hosts below are placeholders; options are as documented for curl 8.7.1.

## Fetch and save {.cols-2 #fetch}
### Download a resource {#download}
| Option | Effect |
| :-- | :-- |
| `-o out.html` | Write the body to a named file |
| `-O` | Name the file after the last URL path segment |
| `-J` | With `-O`, prefer the server's suggested filename |
| `--output-dir dl` | Put saved files in another directory |
| `--create-dirs` | Create missing directories for `-o` paths |
| `--no-clobber` | Never overwrite an existing local file |

### Follow redirects {#redirects}
```sh
curl -L https://docs.example.org/latest
curl -L --max-redirs 3 https://short.example.net/x7
```

`-L` follows `Location:` headers; credentials are not sent to other hosts unless you use `--location-trusted`.

### Quiet but honest output {#quiet}
```sh
curl -sS https://api.example.com/health
curl -# -O https://files.example.com/big.iso
```

`-s` hides progress and errors, `-S` brings errors back. `-#` swaps the meter for a simple bar.

### Resume an interrupted download {#resume}
```sh
curl -C - -O https://files.example.com/big.iso
```

`-C -` makes curl look at the existing local file to work out where to continue.

## Shape the request {.cols-2 #request}
### Method and headers {#headers}
```sh
curl -X DELETE https://api.example.com/items/42
curl -H 'Accept: text/csv' -H 'X-Trace: on' https://api.example.com/report
curl -A 'inventory-bot/2.1' https://shop.example.com/
```

### Send form data {#form-data}
| Option | Body sent |
| :-- | :-- |
| `-d 'qty=3&sku=A7'` | URL-encoded form, POST by default |
| `-d @order.txt` | Read the body from a file |
| `--data-raw '@literal'` | Like `-d`, but `@` has no special meaning |
| `--data-binary @blob.bin` | File contents sent unchanged |
| `--data-urlencode 'note=a & b'` | Value percent-encoded for you |
| `-F 'photo=@cat.jpg'` | `multipart/form-data` with a file part |

### Post JSON {.col-span-2 #post-json}
```sh
curl --json '{"title":"Weekly sync","minutes":30}' https://api.example.com/meetings
curl --json @payload.json https://api.example.com/meetings
jq -n '{title: "Retro"}' | curl --json @- https://api.example.com/meetings
```

`--json` posts the data and sets both `Content-Type` and `Accept` to `application/json`; it does not validate the JSON. See [building JSON with jq](jq.md#construct).

### Query strings {#query}
```sh
curl -G -d 'q=tide tables' -d 'page=2' https://search.example.org/
curl --url-query 'city=Köln' https://weather.example.com/now
```

`-G` turns data options into a GET with a `?` query; `--url-query` encodes like `--data-urlencode`.

### Upload a file {#upload}
```sh
curl -T notes.txt https://dav.example.com/inbox/
curl -T report.pdf ftp://ftp.example.net/incoming/
```

## Inspect responses {.cols-2 #inspect}
### See what happened {#headers-out}
| Option | Shows |
| :-- | :-- |
| `-i` | Response headers followed by the body |
| `-I` | Headers only, via a `HEAD` request |
| `-v` | Request and response headers plus connection details |
| `--trace-ascii trace.log` | Full exchange written to a file, no hex dump |

### Report after the transfer {.col-span-2 #write-out}
```sh
curl -s -o /dev/null -w '%{response_code}\n' https://example.com/
curl -s -o /dev/null -w '%{time_total}s %{size_download} bytes\n' https://example.com/
curl -s -o page.html -w '%{url_effective} %header{content-type}\n' -L https://example.com/
```

Variables use `%{name}`; `%header{name}` prints one response header and `%{json}` dumps every value as JSON.

### Fail on HTTP errors {.tip #fail}
```sh
curl -fsS https://api.example.com/config -o config.json
curl --fail-with-body https://api.example.com/broken
```

Without these flags a `404` still exits with `0`. Both make curl return exit code `22`; `--fail-with-body` also keeps the error page.

## Authentication and TLS {.cols-2 #auth}
### Credentials {#credentials}
| Option | Purpose |
| :-- | :-- |
| `-u alice` | Basic auth; curl prompts for the password |
| `-u alice:secret` | Basic auth with an inline password |
| `-n` | Read credentials from `~/.netrc` |
| `--oauth2-bearer TOKEN` | Send an `Authorization: Bearer` header |

### Keep secrets off the command line {.warning #secrets}
```sh
curl -H @auth-headers.txt https://api.example.com/me
curl --variable %API_TOKEN --expand-oauth2-bearer '{{API_TOKEN}}' https://api.example.com/me
```

Arguments may be visible to other local users in process listings. Headers can be read from a file with `-H @file`, and `--variable %NAME` imports an environment variable.

### Certificates {#certificates}
```sh
curl --cacert internal-ca.pem https://intranet.example.lan/
curl --cert client.pem --key client.key https://mtls.example.com/
```

`-k` skips certificate verification entirely; reserve it for throwaway tests.

### Cookies {#cookies}
```sh
curl -c jar.txt -d 'user=ana' https://app.example.com/login
curl -b jar.txt https://app.example.com/dashboard
```

`-c` saves cookies after the transfer; `-b` sends them from a file or a `name=value` string.

## Robust scripting {.cols-2 #scripting}
### Timeouts and retries {#retries}
| Option | Meaning |
| :-- | :-- |
| `--connect-timeout 5` | Give up if connecting takes longer than 5 s |
| `-m 30` | Cap the whole transfer at 30 s |
| `--retry 4` | Retry on timeouts and HTTP 408, 429, 500, 502, 503, 504 |
| `--retry-all-errors` | With `--retry`, retry on any error |
| `--remove-on-error` | Delete a partial output file when the transfer fails |

### Network routing {#routing}
```sh
curl -x http://proxy.example.lan:3128 https://example.com/
curl --resolve api.example.com:443:192.0.2.10 https://api.example.com/
curl -4 https://example.com/
```

`--resolve` pins a host name to an address without editing hosts files.

### Many transfers at once {#parallel}
```sh
curl -Z --parallel-max 4 -O https://cdn.example.com/a.tgz -O https://cdn.example.com/b.tgz
```

Each URL needs its own `-O` or `-o`. Defaults can live in `~/.curlrc`, one option per line.
