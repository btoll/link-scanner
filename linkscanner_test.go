package linkscanner

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"
)

var (
	linksTpl = `
<!DOCTYPE html>
<head>
	<meta charset="UTF-8">
	<title>Verifier</title>
</head>
<body>
	{{ range . }}
	<li><a href="{{ . }}">{{ . }}></li>
	{{ end }}
</body>
</html>
`
	tpl       = template.Must(template.New("links").Parse(linksTpl))
	target    = "http://example.html"
	goodLinks = []string{
		"http://www.example.com/",
		"http://www.example.com/foobar",
		"http://www.example.com#target",
		"www.example.com/foobar#target",
		"www.example.com/foobar?quux=1&derp=true#target",
		"example.com/foo",
		"example.com/",
		"example.com",
		"/example",
	}
	badLinks = []string{
		"http://wwwexamplecom/",
		"wwwexamplecom/",
		"http://wwwexamplecom#target",
		"wwwexamplecom#target",
		"http://wwwexamplecom?quux=1&derp=true",
		"wwwexamplecom?quux=1&derp=true",
		"wwwexamplecom?quux=1&derp=true#target",
		"http://wwwexamplecom?quux=1&derp=true#target",
		"example",
		"example/",
	}
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (r roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(req)
}

func getClient(links string) *http.Client {
	return &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			// Note that we're hardcoding 200 here b/c all invalid url strings will
			// have been filtered out before this mock is reached (see `parseNodes`).
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(links)),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}),
	}
}

func getScanResults() (*ScanResults, error) {
	var buf bytes.Buffer
	err := tpl.Execute(&buf, slices.Concat(goodLinks, badLinks))
	if err != nil {
		return nil, err
	}
	mockClient := getClient(buf.String())
	return processURLWithClient(context.Background(), target, mockClient)
}

func TestDefaultTagName(t *testing.T) {
	if TagName != "body" {
		t.Errorf("Default tag name should be 'body', got '%s'.\n", TagName)
	}
}

//func TestGetAllLinks(t *testing.T) {
//}

func TestGetFailures(t *testing.T) {
	scanResults, err := getScanResults()
	if err != nil {
		t.Error(err)
	}
	slices.Sort(badLinks)
	slices.Sort(scanResults.Filtered[400])
	if slices.Compare(badLinks, scanResults.Filtered[400]) != 0 {
		t.Errorf("scanResults.GetFailures() %s does not equal the expected failure links %s ", scanResults.Filtered[400], badLinks)
	}
	//	if slices.Compare(scanResults.Filtered[400], scanResults.GetFailures().Filtered[400]) != 0 {
	//		t.Errorf("scanResults.GetFailures() %s does not equal the expected failure links %s ", scanResults.Filtered[400], scanResults.GetFailures().Filtered[400])
	//	}
}

func TestGetSuccesses(t *testing.T) {
	scanResults, err := getScanResults()
	if err != nil {
		t.Error(err)
	}
	slices.Sort(goodLinks)
	slices.Sort(scanResults.LinkResults[200])
	if slices.Compare(goodLinks, scanResults.GetSuccesses().LinkResults[200]) != 0 {
		t.Errorf("scanResults.GetSuccesses() does not equal the expected success links %s ", scanResults.LinkResults[200])
	}
	if slices.Compare(scanResults.LinkResults[200], scanResults.GetSuccesses().LinkResults[200]) != 0 {
		t.Errorf("scanResults.GetSuccesses() does not equal the expected success links %s ", scanResults.LinkResults[200])
	}
}

func TestGetTarget(t *testing.T) {
	scanResults, err := getScanResults()
	if err != nil {
		t.Error(err)
	}
	slices.Sort(goodLinks)
	slices.Sort(scanResults.LinkResults[200])
	if scanResults.Target != target {
		t.Errorf("The target does %s not equal scanResults.Target %s ", target, scanResults.Target)
	}
}

func TestProcessURL(t *testing.T) {
	scanResults, err := getScanResults()
	if err != nil {
		t.Error(err)
	}

	slices.Sort(goodLinks)
	slices.Sort(scanResults.LinkResults[200])

	if slices.Compare(goodLinks, scanResults.LinkResults[200]) != 0 {
		t.Errorf("The returned link results %v do not equal the good links %v.\n", scanResults.LinkResults[200], goodLinks)
	}

	slices.Sort(badLinks)
	slices.Sort(scanResults.Filtered[400])

	if slices.Compare(badLinks, scanResults.Filtered[400]) != 0 {
		t.Errorf("The filtered links %v do not equal the bad links %v.\n", scanResults.Filtered[400], badLinks)
	}
}

func TestSetTagName(t *testing.T) {
	if TagName != "body" {
		t.Errorf("Default tag name should be 'body', got '%s'.\n", TagName)
	}
	SetTagName("foo")
	if TagName != "foo" {
		t.Errorf("Tag name should be 'foo', got '%s'.\n", TagName)
	}
	SetTagName("body")
}

func TestValidateURL(t *testing.T) {
	for _, link := range goodLinks {
		_, err := ValidateURL(link)
		if err != nil {
			t.Error(err)
		}
	}

	for _, link := range badLinks {
		_, err := ValidateURL(link)
		if err == nil {
			t.Error(err)
		}
	}
}
