package main

import (
	"net"
	"net/http"
	"time"
)

var (
	customRoundTripper = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		ResponseHeaderTimeout: 5 * time.Second,
		IdleConnTimeout:       30 * time.Second,
		MaxIdleConns:          60,
		MaxConnsPerHost:       20,
		MaxIdleConnsPerHost:   20,
		TLSHandshakeTimeout:   5 * time.Second,
	}

	client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &MyRoundTripper{
			transport: customRoundTripper,
		},
	}
)

type MyRoundTripper struct {
	transport *http.Transport
}

func (m *MyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Don't mutate the original.  This avoids potential race conditions with concurrent requests.
	// In other words, each call to `RoundTrip` modifies its own copy of `req` when adding headers
	// instead of all of them mutating the original Header struct.
	// No lock management needed.
	req = req.Clone(req.Context())
	addHeaders(req.Header)
	return m.transport.RoundTrip(req)
}

func addHeaders(header http.Header) {
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	header.Set("Accept-Language", "en-US,en;q=0.9")
	// My custom RoundTripper likely isn't handling any decompression where it's likely the default one does.
	// Currently, uncommenting this header will result in any empty links map.
	//	header.Set("Accept-Encoding", "gzip, deflate, br")
	header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
}
