package main

import (
	"github.com/guardian/gocapiclient"
	"github.com/guardian/gocapiclient/queries"
	"github.com/guardian/gocapimodels/content"
	"github.com/guardian/gogridclient"
	"github.com/guardian/gridusagereindex/config"
	"github.com/guardian/gridusagereindex/workers"
	"log"
	"strconv"
)

var appConfig *config.AppConfig

func main() {
	appConfig = config.LoadConfig()

	client := gocapiclient.NewGuardianContentClient(
		appConfig.CapiUrl, appConfig.CapiApiKey)

	searchQueryPaged(client)
}

func createJobs(jobIterator <-chan *content.SearchResponse) <-chan workers.JobResult {
	jobs := make(chan string, 50)
	results := make(chan workers.JobResult)

	defer close(jobs)
	defer close(results)

	usageService := gogridclient.NewUsageService(
		appConfig.UsageUrl, appConfig.GridApiKey)

	for w := 1; w <= 3; w++ {
		go workers.Worker(w, jobs, results, usageService)
	}

	go workers.ResultWorker(results)

	for response := range jobIterator {
		log.Println("Page: " + strconv.FormatInt(int64(response.CurrentPage), 10))
		for _, v := range response.Results {
			jobs <- v.ID
		}
	}

	return results
}

func searchQueryPaged(client *gocapiclient.GuardianContentClient) {
	searchQuery := queries.NewSearchQuery()
	searchQuery.PageOffset = int64(10)

	// TODO: Remove sausages
	showParam := queries.StringParam{"q", "sausages"}
	params := []queries.Param{&showParam}
	searchQuery.Params = params

	iterator := client.SearchQueryIterator(searchQuery)
	results := createJobs(iterator)

	<-results
}
