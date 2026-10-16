package linkscanner

import (
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

var TagName string = "body"

type Links []string
type LinkResults map[int]Links
type ScanResults struct {
	Target      string      `json:"target"`
	Filtered    LinkResults `json:"filtered"`
	LinkResults LinkResults `json:"links"`
}
type LinkMaps struct {
	Filtered    *sync.Map
	LinkResults *sync.Map
}

func (s *ScanResults) GetAllLinks() LinkResults {
	m := LinkResults{}
	maps.Copy(m, s.LinkResults)
	return m
}

func (s *ScanResults) GetFailures() *ScanResults {
	l := &ScanResults{
		LinkResults: make(LinkResults),
	}
	for k, v := range s.LinkResults {
		if !(k >= 200 && k < 400) {
			l.LinkResults[k] = v
		}
	}
	l.Target = s.Target
	return l
}

func (s *ScanResults) GetSuccesses() *ScanResults {
	l := &ScanResults{
		LinkResults: make(LinkResults),
	}
	for k, v := range s.LinkResults {
		if k >= 200 && k < 400 {
			l.LinkResults[k] = v
		}
	}
	l.Target = s.Target
	return l
}

func (s *ScanResults) GetTarget() string {
	return s.Target
}

type Task struct {
	url      string
	resolved string // Used for HEAD request.
	m        *sync.Map
}

func parseNodes(body io.Reader, targetUrl string, tasks chan<- Task, m *LinkMaps) error {
	defer close(tasks)

	node, err := html.Parse(body)
	if err != nil {
		//		logger.Error("html.Parse failed", "err", err)
		return err
	}

	// Get the absolute URI for the links that are just paths, i.e., "/2021/03/18/on-fork-bombs/".
	//
	// We can ignore the error, because if the targetUrl was invalid then
	// client.Get() would have failed in the caller.
	absoluteURL, _ := url.Parse(targetUrl)
	for n := range node.Descendants() {
		if n.Data == TagName {
			for err := range n.Descendants() {
				for _, attr := range err.Attr {
					if attr.Key == "href" {
						// If the url string fails validation we can store it immediately as a bad request.
						// Else, send it through the channel to have a worker send a HEAD request.
						u, err := ValidateURL(attr.Val)
						// We store the original because both `ValidateURL` and `.ResolveReference` will mutate the original string.
						if err != nil {
							actual, loaded := m.Filtered.LoadOrStore(http.StatusBadRequest, []string{attr.Val})
							if loaded {
								a := actual.([]string)
								a = append(a, attr.Val)
								m.Filtered.Store(http.StatusBadRequest, a)
							}
							continue
						}
						resolved := absoluteURL.ResolveReference(u) // This resolves both absolute and relative URLs.
						tasks <- Task{
							url:      attr.Val,
							resolved: resolved.String(),
							m:        m.LinkResults,
						}
					}
				}
			}
			break
		}
	}
	return nil
}

func ProcessURL(ctx context.Context, targetUrl string) (*ScanResults, error) {
	return processURLWithClient(ctx, targetUrl, client)
}

func processURLWithClient(ctx context.Context, targetUrl string, client *http.Client) (*ScanResults, error) {
	scanResults := &ScanResults{
		Target:      targetUrl,
		Filtered:    make(LinkResults),
		LinkResults: make(LinkResults),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetUrl, nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var wgWorker sync.WaitGroup
		m := &LinkMaps{
			Filtered:    &sync.Map{},
			LinkResults: &sync.Map{},
		}
		err := parseNodes(
			resp.Body,
			targetUrl,
			createWorkers(&wgWorker, client),
			m,
		)
		wgWorker.Wait()
		if err != nil {
			return nil, err
		}

		m.LinkResults.Range(func(key, value any) bool {
			uniqueLinks := make(map[string]bool)
			switch v := value.(type) {
			case []string:
				// Deduplicate.
				// TODO: Let's deduplicate before we get here to make it simpler.
				for _, url := range v {
					uniqueLinks[url] = true
				}

			}
			k := key.(int)
			for url := range uniqueLinks {
				scanResults.LinkResults[k] = append(scanResults.LinkResults[k], url)
			}
			return true
		})

		m.Filtered.Range(func(key, value any) bool {
			uniqueLinks := make(map[string]bool)
			switch v := value.(type) {
			case []string:
				// Deduplicate.
				// TODO: Let's deduplicate before we get here to make it simpler.
				for _, url := range v {
					uniqueLinks[url] = true
				}

			}
			k := key.(int)
			for url := range uniqueLinks {
				scanResults.Filtered[k] = append(scanResults.Filtered[k], url)
			}
			return true
		})

		for _, urls := range scanResults.LinkResults {
			slices.Sort(urls)
		}

		for _, urls := range scanResults.Filtered {
			slices.Sort(urls)
		}
	} else {
		return nil, fmt.Errorf("%s return a status of %s, exiting", targetUrl, resp.Status)
	}
	return scanResults, nil
}

func SetTagName(tagName string) {
	TagName = tagName
}

// Just block URLs that aren't valid paths (no host name), host names that have no periods (`.`) in them,
// and ones that have no host name and a querystring and/or a target (/path#target).
// The latter will have its absolute path set to the target, and we don't need to verify that.
// The rest can be passed through and used for HEAD requests.
func ValidateURL(u string) (*url.URL, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	if parsedURL.Host != "" && !strings.Contains(parsedURL.Host, ".") {
		return parsedURL, fmt.Errorf("The URL '%s' does not have a valid host: '%s'.\n", u, parsedURL.Host)
	}
	if parsedURL.Host == "" && !strings.Contains(parsedURL.Path, ".") {
		if strings.Index(parsedURL.Path, "/") == 0 {
			return parsedURL, nil
		}
		return parsedURL, fmt.Errorf("The URL '%s' does not have a valid path: '%s'.\n", u, parsedURL.Path)
	}
	return parsedURL, nil
}
