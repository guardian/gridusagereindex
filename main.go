package main

import (
	"github.com/guardian/gobby"
	"github.com/guardian/gocapiclient"
	"github.com/guardian/gocapiclient/queries"
	"github.com/guardian/gogridclient"
	"github.com/guardian/gridusagereindex/config"
	"github.com/guardian/gridusagereindex/workers"
	"log"
	"time"
)

var appConfig *config.AppConfig

func main() {
	appConfig = config.LoadConfig()

	client := gocapiclient.NewGuardianContentClient(
		appConfig.CapiUrl, appConfig.CapiApiKey)

	gobdb := gobby.New(appConfig.GobbyFile)
	gobdb.Load()

	searchQueryPaged(client, gobdb)
}

func createJobs(jobIterator <-chan *queries.SearchPageResponse, gobdb *gobby.Gobby) <-chan workers.JobResult {
	jobs := make(chan string, 50)
	results := make(chan workers.JobResult)

	defer close(jobs)
	defer close(results)

	usageService := gogridclient.NewUsageService(
		appConfig.UsageUrl, appConfig.GridApiKey)

	for w := 1; w <= 3; w++ {
		go workers.ReindexWorker(w, jobs, results, usageService, gobdb)
	}

	go workers.ResultWorker(results)

	for page := range jobIterator {
		if page.Err != nil {
			log.Fatal(page.Err)
		}

		for _, v := range page.SearchResponse.Results {
			jobs <- v.ID
		}

		gobdb.Save()
	}

	return results
}

func searchQueryPaged(client *gocapiclient.GuardianContentClient, gobdb *gobby.Gobby) {
	searchQuery := queries.NewSearchQuery()
	searchQuery.PageOffset = int64(10)

	fromDate := appConfig.FromDate
	toDate := appConfig.ToDate

	dateLayout := "2006-01-02"

	var err error
	var from, to time.Time

	from, err = time.Parse(dateLayout, fromDate)
	to, err = time.Parse(dateLayout, toDate)

	log.Println("Starting at:", from)
	log.Println("Ending at:", to)

	if err != nil {
		log.Fatal(err)
	}

	var nextDay time.Time
	var fromString string
	var toString string

	for from.Before(to) {
		nextDay = from.Add(time.Hour * 24)

		fromString = from.Format(dateLayout)
		toString = nextDay.Format(dateLayout)

		gobbyJobId := fromString + "-" + toString + "-capi-usage-reindex"
		_, exists := gobdb.Get(gobbyJobId)

		if exists {
			log.Println("Already seen:", fromString, "-", toString, "(skipping)")
			from = from.Add(time.Hour * 24)
			continue
		}

		fromParam := queries.StringParam{"from-date", fromString}
		toParam := queries.StringParam{"to-date", toString}

		log.Println("Currently working on:", fromString, "-", toString)

		params := []queries.Param{&fromParam, &toParam}
		searchQuery.Params = params

		iterator := client.SearchQueryIterator(searchQuery)
		results := createJobs(iterator, gobdb)

		<-results

		jobStatus := gobby.JobStatus{gobbyJobId, "done", nil}

		gobdb.Set(gobbyJobId, jobStatus)
		gobdb.Save()

		from = from.Add(time.Hour * 24)
	}
}
