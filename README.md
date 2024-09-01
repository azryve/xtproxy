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

* Serves files simultaneously with FTP/TFTP/HTTP.
* Sources files from S3 bucket/HTTP file share/local directory.
* Can combine multiple sources of files.
* Supports IPv4/IPv6.

## Known Limitations

* No client authentication.
* Limited testing.
* No upload capability.

## Usage

### s3 bucket

```
export XTPROXY_S3_CREDENTIALS="ACCESSKEYID:secretaccesskeyvalue"

./xtproxy "s3://s3.amazonaws.com/eu-north-1/myownbucket /"
```

### s3 bucket + local dir

```

export XTPROXY_S3_CREDENTIALS="ACCESSKEYID:secretaccesskeyvalue"

./xtproxy \
  "s3://s3.amazonaws.com/eu-north-1/myownbucket /" \
  "file:///var/spool/localfileshare /localfileshare"
```
