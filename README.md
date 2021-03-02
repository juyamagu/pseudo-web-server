# pseudo-web-server

## Usage

```bash
# Usage
$ ./pseudo-web-server -help
Usage of ./pseuudo-web-server:
  -length-max int
        Maximum body size in for randomly generating response body. (default 1048576)
  -length-unit int
        Default number of bytes to be written in a single loop (default 1024)
  -listen-addr string
        Server listen address. (default ":8080")
  -time-max int
        Maximum time in [sec] for randomly determining the processing time. (default 600)

# Launch
$ ./pseudo-web-server

# Launch with options
$ sudo ./pseudo-web-server -listen-addr ":80"

# Client:
# You can utilize GET Query to adjust the server response:
#     ?length="Content-Length (in bytes)"
#     ?unit="processing unit (in bytes)"
#     ?time="response time (in seconds)"
$ curl "localhost:8080/?length=<content-length>&unit=<processing-unit>&time=<response-time>"
```

