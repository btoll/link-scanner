package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

var (
	numWorkers int
	tagName    string
	targetUrl  string
	quiet      bool

	// TODO: Put logging behind a -debug or -verbose flag.
	logger *slog.Logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
)

type Links map[int][]string
type Target map[string]any
type Targets []Target

type Task struct {
	url string
	m   *sync.Map
}

//func boolToInt(b bool) int {
//	if !b {
//		return 0
//	}
//	return 1
//}

func createWorkers(wg *sync.WaitGroup, numWorkers int) chan Task {
	tasks := make(chan Task, numWorkers)
	for i := 0; i < numWorkers; i += 1 {
		wg.Go(func() {
			worker(tasks)
		})
	}
	return tasks
}

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

func parseNodes(body io.Reader, tasks chan<- Task, m *sync.Map) {
	node, err := html.Parse(body)
	if err != nil {
		close(tasks)
		logger.Error("html.Parse failed", "err", err)
		os.Exit(1)
	}
	for n := range node.Descendants() {
		if n.Data == tagName {
			for e := range n.Descendants() {
				for _, a := range e.Attr {
					if a.Key == "href" {
						// Validate that it's well-formed.
						if u, err := url.Parse(a.Val); err == nil && u.Scheme != "" {
							tasks <- Task{
								url: a.Val,
								m:   m,
							}
						}
					}
				}
			}
			close(tasks)
			break
		}
	}
}

func processURL(url string) (Target, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	target := make(Target)
	target["target"] = url

	if resp.StatusCode == 200 {
		var wgWorker sync.WaitGroup
		m := &sync.Map{}
		parseNodes(
			resp.Body,
			createWorkers(&wgWorker, numWorkers),
			m,
		)
		wgWorker.Wait()

		links := make(Links)
		m.Range(func(key, value any) bool {
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
				links[k] = append(links[k], url)
			}
			return true
		})

		for _, urls := range links {
			slices.Sort(urls)
		}
		target["links"] = links
	} else {
		target["links"] = struct{}{}
	}
	return target, nil
}

func worker(tasks <-chan Task) {
	for task := range tasks {
		// TODO: Allow HTTP verb (HEAD, GET) to be a CLI param.
		req, err := http.NewRequest(http.MethodHead, task.url, nil)
		if err != nil {
			// TODO: Let's have a meaningful error status code here.
			task.m.Store(-1, []string{task.url})
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			// TODO: Let's have a meaningful error status code here.
			task.m.Store(-2, []string{task.url})
			continue
		}
		actual, loaded := task.m.LoadOrStore(resp.StatusCode, []string{task.url})
		// TODO: This could still be subject to a race condition.
		if loaded {
			// The key already existed so append to the existing slice and store.
			a := actual.([]string)
			a = append(a, task.url)
			task.m.Store(resp.StatusCode, a)
		}
	}
}

func main() {
	flag.StringVar(&tagName, "tagName", "body", "The HTML node to target in which to get the links")
	flag.StringVar(&targetUrl, "url", "", "The URL to check for valid links.")
	flag.IntVar(&numWorkers, "w", 3, "The number of workers in the worker pool.")
	flag.IntVar(&numWorkers, "workers", 20, "The number of workers in the worker pool.")
	flag.BoolVar(&quiet, "q", false, "Suppress output")
	flag.BoolVar(&quiet, "quiet", false, "Suppress output")
	flag.Parse()

	var wgMain sync.WaitGroup
	allURLs := getURLs(targetUrl)
	targets := make(Targets, len(allURLs))

	for i, url := range allURLs {
		wgMain.Go(func() {
			target, err := processURL(url)
			if err != nil {
				logger.Error(err.Error())
				return
			}
			targets[i] = target
		})
	}

	wgMain.Wait()

	if !quiet {
		b, err := json.Marshal(targets)
		if err != nil {
			logger.Error(err.Error())
		}
		fmt.Println(string(b))
	}
}
