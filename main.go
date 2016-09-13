package main

import (
	"github.com/guardian/gobby"
	"github.com/guardian/gocapiclient"
	"github.com/guardian/gocapiclient/queries"
	"github.com/guardian/gogridclient"
	"github.com/guardian/gridusagereindex/config"
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

func searchQueryPaged(client *gocapiclient.GuardianContentClient, gobdb *gobby.Gobby) {
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
		searchQuery := queries.NewSearchQuery()
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
		usageService := gogridclient.NewUsageService(
			appConfig.UsageUrl, appConfig.GridApiKey)

		for page := range iterator {
			if page.Err != nil {
				log.Println("Could not process page!", page.Err)
				continue
			}

			for _, v := range page.SearchResponse.Results {
				log.Println("Processing job", v.ID)

				_, exists := gobdb.Get(v.ID)

				if exists {
					log.Println("Already seen", v.ID, "(skipping)")
					continue
				}

				response, err := usageService.Reindex(v.ID)
				jobStatus := gobby.JobStatus{v.ID, response.Status, nil}

				gobdb.Set(v.ID, jobStatus)

				if err != nil {
					log.Println("Error reading from CAPI", err)
				}

				log.Println("Done:", v.ID, "::", response.Status)
			}

			gobdb.Save()
		}

		jobStatus := gobby.JobStatus{gobbyJobId, "done", nil}

		gobdb.Set(gobbyJobId, jobStatus)
		gobdb.Save()

		from = from.Add(time.Hour * 24)
	}
}
