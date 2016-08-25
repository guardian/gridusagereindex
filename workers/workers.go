package workers

import (
	"github.com/guardian/gogridclient"
	"log"
)

type JobResult struct {
	Status string
	Id     string
}

func Worker(id int, jobs <-chan string, results chan<- JobResult, usageService *gogridclient.UsageService) {
	for j := range jobs {
		log.Println("worker", id, "processing job", j)

		response, err := usageService.Reindex(j)
		if err != nil {
			log.Fatal(err)
		}

		results <- JobResult{response.Status, j}
	}
}

func ResultWorker(results <-chan JobResult) {
	for msg := range results {
		log.Println("Done: " + msg.Id + ": " + msg.Status)
	}
}
