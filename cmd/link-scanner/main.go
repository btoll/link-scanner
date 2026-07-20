package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"sync"

	linkscanner "github.com/btoll/link-scanner"
)

var (
	tagName   string
	targetUrl string
	quiet     bool

	// TODO: Put logging behind a -debug or -verbose flag.
//	logger *slog.Logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
)

func main() {
	flag.StringVar(&tagName, "tagName", "body", "The HTML node to target in which to get the links")
	flag.StringVar(&targetUrl, "url", "", "The URL to check for valid links.")
	flag.BoolVar(&quiet, "q", false, "Suppress output")
	flag.BoolVar(&quiet, "quiet", false, "Suppress output")
	flag.Parse()

	linkscanner.SetTagName(tagName)
	allURLs := linkscanner.GetURLs(targetUrl)
	targets := make([]*linkscanner.ScanResults, len(allURLs))

	var wgMain sync.WaitGroup
	for i, url := range allURLs {
		wgMain.Go(func() {
			target, err := linkscanner.ProcessURL(url)
			if err != nil {
				//				logger.Error(err.Error())
				return
			}
			targets[i] = target
		})
	}
	wgMain.Wait()

	if !quiet {
		var b []byte
		var err error
		if len(allURLs) < 2 {
			b, err = json.Marshal(targets[0])
			if err != nil {
				//			logger.Error(err.Error())
			}
		} else {
			b, err = json.Marshal(targets)
			if err != nil {
				//			logger.Error(err.Error())
			}
		}
		fmt.Println(string(b))
	}
}
