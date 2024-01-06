package main

import (
	"log"
	"net/http"

	"github.com/mendelgusmao/scoredb/endpoints"
	"github.com/mendelgusmao/scoredb/lib/database/persistence"
	"github.com/mendelgusmao/scoredb/middleware"
)

func main() {
	var (
		err     error
		persist *persistence.Persistence
	)

	if err = readConfig(); err != nil {
		log.Fatal(err)
	}

	if ScoreDB.SnapshotPath != "" {
		persist = persistence.NewPersistence(
			endpoints.DB,
			persistence.Configuration{
				SnapshotPath:     ScoreDB.SnapshotPath,
				SnapshotInterval: ScoreDB.SnapshotInterval,
				SnapshotWaitLoad: ScoreDB.SnapshotWaitLoad,
			},
		)

		persist.Load()
		persist.Work()
	}

	log.Println("starting scoredb server at", ScoreDB.Listen)

	router := middleware.NewLoading(endpoints.Router, ScoreDB.SnapshotWaitLoad, persist)

	if ScoreDB.Logging {
		loggingRouter := middleware.NewLogger(router)
		err = http.ListenAndServe(ScoreDB.Listen, loggingRouter)
	} else {
		err = http.ListenAndServe(ScoreDB.Listen, router)
	}

	log.Fatal(err)
}
