![Go](https://github.com/jeffrom/tulpa/workflows/Go/badge.svg)

# tulpa

tulpa is a command line utility for live reloading applications. tulpa serves
your application through a proxy. Anytime the proxy receives an HTTP request, it
will automatically rerun your command if the filesystem has changed.

`tulpa` is language agnostic - it can be configured to work with anything: Go,
Rust, Ruby, Python, scripts, and arbitrary commands.

tulpa is a fork of [tychus](https://github.com/devlocker/tychus). tulpa adds
some output formatting, debounce, and other bells and whistles. Thanks to
[@PatKoperwas](https://github.com/PatKoperwas) for the great work!

## Installation

### With Go
Assuming you have a working Go environment and `GOPATH/bin` is in your `PATH`

```
go get github.com/jeffrom/tulpa
```

### Windows
Currently isn't supported :(

## Usage

Usage is simple, `tulpa [command]` A proxy will be started on port `4000`. When
an HTTP request comes in and the filesystem has changed, your command will be
rerun.

```
tulpa go run main.go
```

## Options
tulpa has a few options. In most cases the defaults should be sufficient.

```yaml
  -a, --app-port int         port your application runs on, overwritten by ENV['PORT'] (default 3000)
  -p, --proxy-port int       proxy port (default 4000)
  -x, --ignore string        comma separated list of directories to ignore file changes in. (default node_modules,log,tmp,vendor)
  -w, --wait                 Wait for command to finish before proxying a request.
  -t, --timeout int          timeout for proxied requests (default 10)

  -h, --help                 help for tulpa
      --debug                print debug output
      --version              version for tulpa
```

Note: tulpa will not look for file system changes in any hidden directories
(those beginning with `.`).

## Examples

**Example: Web Servers**

```
// Go - Hello World Server
$ tulpa go run main.go
[tulpa] Proxing requests on port 4000 to 3000
[Go App] App Starting on Port 3000

// Make a request
$ curl localhost:4000
Hello World
$ curl localhost:4000
Hello World

// Save a file, next request will restart your webapp
$ curl localhost:4000
[Go App] App Starting on Port 3000
Hello World
```

This can work with any webserver:

```
// Rust
tulpa cargo run

// Ruby
tulpa ruby myapp.rb
```

Need to pass flags? Stick the command in quotes

```
tulpa "ruby myapp.rb -e development"
```

Complicated command? Stick it in quotes

```
tulpa "go build -o my-bin && echo 'Built Binary' && ./my-bin"
```

**Example: Scripts + Commands**

Scenario: You have a webserver running on port `3005`, and it serves static
files from the `/public` directory. In the `/docs` folder are some markdown
files. Should they change, you want them rebuilt and placed into the `public`
directory so the server can pick them up.

```
tulpa "multimarkdown docs/index.md > public/index.html" --wait --app-port=3005
```

Now, when you make a request to the proxy on `localhost:4000`, `tulpa` will
pause the request (that's what the `--wait` flag is for) until `multimarkdown`
finishes. Then request will be forwarded to the server on port `3005`.
`multimarkdown` will only be run if the filesystem has changed.

**Other Proxy Goodies**

**Error messages**

If you make a syntax error, or your program won't build for some reason, the
stderr output will be returned by the proxy. Handy for the times you can't see
you server (its in another pane / tab / tmux split).
# thanks

Thanks to [@PatKoperwas](https://github.com/PatKoperwas) for tychus, for which
this program is originally a fork.
