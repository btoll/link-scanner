package linkscanner

import (
	"net/http"
	"sync"
)

// We might want to change this at some point so concurrent calls won't overwrite this.
// For now, nothing is doing that, so we'll keep it simple.
var NumWorkers int = 10

func worker(tasks <-chan Task, client *http.Client) {
	for task := range tasks {
		// TODO: Allow HTTP verb (HEAD, GET) to be a CLI param.
		// The resolved url string is used for the HEAD request and then discarded.
		req, err := http.NewRequest(http.MethodHead, task.resolved, nil)
		if err != nil {
			// TODO: Let's have a meaningful error status code here.
			task.m.Store(-1, []string{task.url})
			continue
		}
		var statusCode int
		resp, err := client.Do(req)
		if err != nil {
			// TODO: Let's have a meaningful error status code here.
			statusCode = 700
		} else {
			statusCode = resp.StatusCode
			resp.Body.Close()
		}
		if statusCode == 403 {
			req, _ := http.NewRequest(http.MethodGet, task.resolved, nil)
			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
			}
			statusCode = resp.StatusCode
		}
		// The original string is stored and outputted.
		actual, loaded := task.m.LoadOrStore(statusCode, []string{task.url})
		// TODO: This could still be subject to a race condition.
		if loaded {
			a := actual.([]string)
			a = append(a, task.url)
			task.m.Store(statusCode, a)
		}
	}
}

func createWorkers(wg *sync.WaitGroup, client *http.Client) chan Task {
	tasks := make(chan Task, NumWorkers)
	for i := 0; i < NumWorkers; i += 1 {
		wg.Go(func() {
			worker(tasks, client)
		})
	}
	return tasks
}
