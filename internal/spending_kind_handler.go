package internal

import (
	"net/http"

	"github.com/gorilla/mux"
)

type SpendKindHandler struct {
	db                  SpenderDB
	loginSessionHandler *LoginSessionManager
}

func SpendKindHandlerSetup(router *mux.Router, db SpenderDB, loginSessionManager *LoginSessionManager) {
	handler := &SpendKindHandler{
		db:                  db,
		loginSessionHandler: loginSessionManager,
	}

	router.HandleFunc("", handler.handleGetDefSpendKinds)
	router.HandleFunc("/{username}", handler.handleGetSpendKinds)
}

func (handler *SpendKindHandler) handleGetDefSpendKinds(w http.ResponseWriter, r *http.Request) {
	//TODO: check logged

	spKinds, err := handler.db.GetAllDefaultSpendKinds()
	if err != nil {
		SendAPIErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	SendAPIOKRespWithData(w, "success", spKinds)
}

func (handler *SpendKindHandler) handleGetSpendKinds(w http.ResponseWriter, r *http.Request) {
	//TODO: check logged

	vars := mux.Vars(r)
	username := vars["username"]
	spKinds, err := handler.db.GetSpendKinds(username)
	if err != nil {
		SendAPIErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	SendAPIOKRespWithData(w, "success", spKinds)
}
