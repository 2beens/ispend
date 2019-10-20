package ispend

import (
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
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
		sendErr := SendAPIErrorResp(w, err.Error(), http.StatusBadRequest)
		if sendErr != nil {
			log.Errorf("error while sending error response to client [get def spend kinds]: %s", sendErr.Error())
		}
		return
	}
	err = SendAPIOKRespWithData(w, "success", spKinds)
	if err != nil {
		log.Errorf("failed to send response to client [get DEF SPEND KINDS]. details: %s", err.Error())
	}
}

func (handler *SpendKindHandler) handleGetSpendKinds(w http.ResponseWriter, r *http.Request) {
	//TODO: check logged

	vars := mux.Vars(r)
	username := vars["username"]
	spKinds, err := handler.db.GetSpendKinds(username)
	if err != nil {
		sendErr := SendAPIErrorResp(w, err.Error(), http.StatusBadRequest)
		if sendErr != nil {
			log.Errorf("error while sending error response to client [get spend kinds (%s)]: %s", username, sendErr.Error())
		}
		return
	}
	err = SendAPIOKRespWithData(w, "success", spKinds)
	if err != nil {
		log.Errorf("failed to send response to client [get spend kinds (%s)]. details: %s", username, err.Error())
	}
}
