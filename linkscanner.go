package linkscanner

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

var TagName string = "body"

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

func parseNodes(body io.Reader, tasks chan<- Task, m *sync.Map) {
	node, err := html.Parse(body)
	if err != nil {
		close(tasks)
		//		logger.Error("html.Parse failed", "err", err)
		os.Exit(1)
	}
	for n := range node.Descendants() {
		if n.Data == TagName {
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

func GetURLs(url string) []string {
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

func ProcessURL(url string) (Target, error) {
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
			CreateWorkers(&wgWorker),
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
