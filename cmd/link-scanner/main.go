package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
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

func getFileInput(arg string) (io.ReadCloser, error) {
	after, found := strings.CutPrefix(arg, "/dev/fd/")
	if found {
		if fd, err := strconv.Atoi(after); err == nil {
			file := os.NewFile(uintptr(fd), "")
			if _, err = file.Stat(); err == nil {
				return file, nil
			}
		}
	}
	if arg == "-" {
		return io.NopCloser(os.Stdin), nil
	}
	return os.Open(arg)
}

func getURLs(url string) []string {
	var allURLs []string
	if url != "" {
		allURLs = []string{url}
	} else if len(os.Args) > 1 {
		reader, err := getFileInput(flag.Args()[0])
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			allURLs = append(allURLs, scanner.Text())
		}
	}
	return allURLs
}

func main() {
	flag.StringVar(&tagName, "tagName", "body", "The HTML node to target in which to get the links")
	flag.StringVar(&targetUrl, "url", "", "The URL to check for valid links.")
	flag.BoolVar(&quiet, "q", false, "Suppress output")
	flag.BoolVar(&quiet, "quiet", false, "Suppress output")
	flag.Parse()

	linkscanner.SetTagName(tagName)
	allURLs := getURLs(targetUrl)
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
