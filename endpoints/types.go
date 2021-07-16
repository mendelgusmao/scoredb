package endpoints

import (
	"github.com/julienschmidt/httprouter"
	"github.com/mendelgusmao/scoredb/lib/database"
)

var (
	Router = httprouter.New()
	db     = database.NewDatabase()
)

type createRequest struct {
	database.Configuration
	updateRequest
}

type updateRequest struct {
	Documents []database.Document `json:"documents"`
}
