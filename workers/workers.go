package workers

import (
	"github.com/guardian/gobby"
	"github.com/guardian/gogridclient"
	"log"
)

type JobResult struct {
	Status string
	Id     string
}

func ReindexWorker(
	id int,
	jobs <-chan string,
	results chan<- JobResult,
	usageService *gogridclient.UsageService,
	gobdb *gobby.Gobby,
) {
	for j := range jobs {
		log.Println("Worker", id, "processing job", j)

		_, exists := gobdb.Get(j)

		if exists {
			log.Println("Already seen", j, "(skipping)")
			continue
		}

		response, err := usageService.Reindex(j)
		jobStatus := gobby.JobStatus{j, response.Status, nil}

		gobdb.Set(j, jobStatus)

		if err != nil {
			log.Fatal(err)
		}

		results <- JobResult{response.Status, j}
	}
}

func ResultWorker(results <-chan JobResult) {
	for msg := range results {
		log.Println("Done:", msg.Id, "::", msg.Status)
	}
}
