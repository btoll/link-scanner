package linkscanner

import (
	"net/http"
	"sync"
)

// We might want to change this at some point so concurrent calls won't overwrite this.
// For now, nothing is doing that, so we'll keep it simple.
var NumWorkers int = 10

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

func CreateWorkers(wg *sync.WaitGroup) chan Task {
	tasks := make(chan Task, NumWorkers)
	for i := 0; i < NumWorkers; i += 1 {
		wg.Go(func() {
			worker(tasks)
		})
	}
	return tasks
}
