package endpoints

import (
	"github.com/julienschmidt/httprouter"
	"github.com/mendelgusmao/scoredb/lib/database"
)

var (
	Router = httprouter.New()
	DB     = database.NewDatabase[any]()
)

type createRequest struct {
	database.FuzzySetConfiguration
	updateRequest
}

type updateRequest struct {
	Documents []database.Document[any] `json:"documents"`
}
