package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func init() {
	endpoint := "/collection/:collectionName"

	Router.PUT(endpoint, create)
	Router.GET(endpoint, query)
	Router.PATCH(endpoint, update)
	Router.DELETE(endpoint, remove)
}

func create(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	collectionName := params.ByName("collectionName")
	request := createRequest{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}

	err := DB.CreateCollection(collectionName, request.FuzzySetConfiguration, request.Documents)

	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func query(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	collectionName := params.ByName("collectionName")
	key := r.FormValue("key")

	entries, err := DB.Query(collectionName, key)

	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entries)
}

func update(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	collectionName := params.ByName("collectionName")
	request := updateRequest{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := DB.UpdateCollection(collectionName, request.Documents); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func snapshot(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	w.WriteHeader(http.StatusOK)
	w.Write(DB.Snapshot())
}

func remove(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	collectionName := params.ByName("collectionName")

	if err := DB.RemoveCollection(collectionName); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeError(w http.ResponseWriter, httpStatus int, err error) {
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
