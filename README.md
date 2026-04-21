# link-scanner

```
Usage of link-scanner:
  -q    Suppress output
  -quiet
        Suppress output
  -tagName string
        The HTML node to target in which to get the links (default "body")
  -url string
        The URL to check for valid links.
  -w int
        The number of workers in the worker pool. (default 3)
  -workers int
        The number of workers in the worker pool. (default 3)
```

## Examples

The link(s) can be passed to the binary as a CLI parameter, from `stdin` or as a file descriptor.

As a CLI parameter:

```bash
$ ./link-scanner -url https://go.dev/ref/mem | jq
{
  "200": [
    "http://www.google.com/intl/en/policies/privacy/",
    "https://github.com/golang",
    "https://github.com/golang/go/issues",
    "https://golangweekly.com/",
    "https://google.com",
    "https://groups.google.com/g/golang-nuts",
    "https://hachyderm.io/@golang",
    "https://invite.slack.golangbridge.org/",
    "https://pkg.go.dev",
    "https://pkg.go.dev/about",
    "https://pkg.go.dev/std",
    "https://policies.google.com/technologies/cookies",
    "https://reddit.com/r/golang",
    "https://stackoverflow.com/questions/tagged/go?tab=Newest",
    "https://stackoverflow.com/tags/go",
    "https://www.meetup.com/pro/go",
    "https://www.reddit.com/r/golang/"
  ],
  "403": [
    "https://dl.acm.org/doi/10.1145/1375581.1375591",
    "https://twitter.com/golang",
    "https://www.twitter.com/golang"
  ],
  "404": [
    "https://bsky.app/profile/golang.org"
  ]
}
```

> Note that some servers with bot detection will reject `HEAD` requests, which is what this tool is using.  In the above example, the links in the `403` and `404` lists are being blocked or rejected by sophisticated user agent detectors that often flag `HEAD` requests.

From `stdin`:

```bash
$ cat << EOF | ./link-scanner . - | jq
https://internals-for-interns.com/posts/go-runtime-scheduler/
https://go.dev/ref/mem
https://benjamintoll.com/2022/08/15/on-testing-website-links/
EOF
```

As a file descriptor:

```bash
$ ./link-scanner . <(cat links.txt) | jq
```

