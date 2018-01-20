package workers

import (
	"net/http"
	"sync"
	"time"

	. "github.com/dpecos/golinks/common"
)

func Start(nWorkers int, linkJobs chan Link, results chan Link) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(nWorkers)

	for i := 0; i < nWorkers; i++ {
		go linkChecker(linkJobs, results, &wg)
	}

	return &wg
}

func Stop(linkJobs chan Link, results chan Link, wg *sync.WaitGroup, wgRes *sync.WaitGroup) {
	close(linkJobs)
	wg.Wait()

	close(results)
	wgRes.Wait()
}

func linkChecker(linkJobs chan Link, results chan Link, wg *sync.WaitGroup) {
	defer wg.Done()

	var netClient = &http.Client{
		Timeout: time.Second * 5,
	}

	for link := range linkJobs {
		resp, err := netClient.Get(link.Url)
		if err != nil || resp.StatusCode < 200 && resp.StatusCode >= 300 {
			if resp != nil {
				link.Status = resp.StatusCode
			} else {
				link.Status = -1
			}
		} else {
			link.Status = resp.StatusCode
		}

		results <- link
	}
}
