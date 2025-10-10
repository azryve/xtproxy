# xtproxy

xtproxy is an FTP/TFTP/HTTP file server proxy.

## Overview

Originally, xtproxy was intended to serve image files for network switches.

It helps when:

* You can't spin up a file server in a directly connected network.
* Files are too big to serve from the local file system.
* TFTP from a remote source is too slow and unreliable, but clients are too limited for FTP/HTTP.
* Clients are smart enough for FTP, but your firewall is not, or you have no control over it.
* You have an IPv6-only segment, but the client does not support it without an upgrade.

## Features

* Serves files simultaneously with FTP/TFTP/HTTP/Webdav.
* Sources files from S3 bucket/HTTP/Webdav file share/local directory.
* Can combine multiple sources of files.
* Supports IPv4/IPv6.

## Known Limitations

* No client authentication.
* No directory listing.
* Limited testing.
* No upload capability.

## Usage

### Configuration params


| Flag                  | Type          | Description                                                                                                                  | Env var                  | Default                          |
| --------------------- | ------------- | ---------------------------------------------------------------------------------------------------------------------------- | ------------------------ | -------------------------------- |
| `--debug`             | `bool`        | Enable debugging                                                                                                             | `XTPROXY_DEBUG`          | `false`                          |
| `--port-ftp`          | `int`         | FTP TCP control port                                                                                                         | `XTPROXY_PORT_FTP`       | `21`                             |
| `--port-http`         | `int`         | HTTP TCP port                                                                                                                | `XTPROXY_PORT_HTTP`      | `80`                             |
| `--port-tftp`         | `int`         | TFTP UDP port                                                                                                                | `XTPROXY_PORT_TFTP`      | `69`                             |
| `-i, --ifaces-listen` | `[]string` | Listen on addresses from specific interfaces                                                                                    | `XTPROXY_IFACES_LISTEN`  | *(none, can be provided multiple times)* |
| `--webdav-handle`     | `string`      | WebDAV handle for HTTP server                                                                                                | `XTPROXY_WEBDAV_HANDLE`  | `"/.webdav"`                     |
|  *(env only)*         | —             | Secret credentials for S3 access                                                                                             | `XTPROXY_S3_CREDENTIALS` | *(none; required if S3 is used)* |

xtproxy can be configured both by envs and cli arguments.
In case when both are provided cli args take precedence over envs.

### Examples

Listen only on addresses from specific interfaces
```
# via cli
./xtproxy -i eth0 -i eth1 "file:///tmp /tmp"

# via env
export XTPROXY_IFACES_LISTEN="eth0 eth1"
./xtproxy "file:///tmp /tmp"
```

Disable ftp/tftp
```
export XTPROXY_PORT_TFTP=0
export XTPROXY_PORT_FTP=0
./xtproxy "file:///tmp /tmp"
```

#### s3 bucket

```
export XTPROXY_S3_CREDENTIALS="ACCESSKEYID:secretaccesskeyvalue"

./xtproxy "s3://s3.amazonaws.com/eu-north-1/myownbucket /"
```

#### s3 bucket + local dir

```

export XTPROXY_S3_CREDENTIALS="ACCESSKEYID:secretaccesskeyvalue"

./xtproxy \
  "s3://s3.amazonaws.com/eu-north-1/myownbucket /" \
  "file:///var/spool/localfileshare /localfileshare"
```

#### daisy chain xtproxy -> xtproxy via webdav

Sometimes its useful to have an additional proxy layer.
It can be achived by daisy chaining two instances via webdav.

```
# serves data from s3 and runs in aws
user@aws-vm1:~$ xtproxy "s3://s3.amazonaws.com/eu-north-1/myownbucket /"
```

```
# runs on remote site and serves data from aws-vm proxy
user@remotesite1:~$ xtproxy "webdav://aws-vm1.example.com/.webdav /"
```
