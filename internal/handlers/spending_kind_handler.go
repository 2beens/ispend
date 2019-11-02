package handlers

import (
	"net/http"

	"github.com/2beens/ispend/internal/db"
	"github.com/2beens/ispend/internal/platform"
	"github.com/gorilla/mux"
)

type SpendKindHandler struct {
	db                  db.SpenderDB
	loginSessionHandler *platform.LoginSessionManager
}

func SpendKindHandlerSetup(router *mux.Router, db db.SpenderDB, loginSessionManager *platform.LoginSessionManager) {
	handler := &SpendKindHandler{
		db:                  db,
		loginSessionHandler: loginSessionManager,
	}

	router.HandleFunc("", handler.handleGetDefSpendKinds).Methods("GET")
	router.HandleFunc("/{username}", handler.handleGetSpendKinds).Methods("GET")
}

func (handler *SpendKindHandler) handleGetDefSpendKinds(w http.ResponseWriter, r *http.Request) {
	//TODO: check logged

	//TODO: don't go directly to DB

	spKinds, err := handler.db.GetAllDefaultSpendKinds()
	if err != nil {
		platform.SendAPIErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	platform.SendAPIOKRespWithData(w, "success", spKinds)
}

func (handler *SpendKindHandler) handleGetSpendKinds(w http.ResponseWriter, r *http.Request) {
	//TODO: check logged

	vars := mux.Vars(r)
	username := vars["username"]
	spKinds, err := handler.db.GetSpendKinds(username)
	if err != nil {
		platform.SendAPIErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	platform.SendAPIOKRespWithData(w, "success", spKinds)
}
