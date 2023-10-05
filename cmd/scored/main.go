package main

import (
	"log"
	"net/http"

	"github.com/mendelgusmao/scoredb/endpoints"
	"github.com/mendelgusmao/scoredb/lib/database/persistence"
	"github.com/mendelgusmao/scoredb/middleware"
)

func main() {
	var err error

	if err = readConfig(); err != nil {
		log.Fatal(err)
	}

	if ScoreDB.SnapshotPath != "" {
		persistence := persistence.NewPersistence(endpoints.DB, persistence.Configuration{
			SnapshotPath:     ScoreDB.SnapshotPath,
			SnapshotInterval: ScoreDB.SnapshotInterval,
		})

		go persistence.Load()
		go persistence.Work()
	}

	log.Println("starting scoredb server at", ScoreDB.Listen)

	if ScoreDB.Logging {
		loggingRouter := middleware.NewLogger(endpoints.Router)
		err = http.ListenAndServe(ScoreDB.Listen, loggingRouter)
	} else {
		err = http.ListenAndServe(ScoreDB.Listen, endpoints.Router)
	}

	log.Fatal(err)
}
