package ispend

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type SpendKindHandler struct {
	db                  SpenderDB
	loginSessionHandler *LoginSessionManager
}

func NewSpendKindHandler(db SpenderDB, loginSessionManager *LoginSessionManager) *SpendingHandler {
	return &SpendingHandler{
		db:                  db,
		loginSessionManager: loginSessionManager,
	}
}

func (handler *SpendKindHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if r.URL.Path == "/spending/kind" {
			handler.handleGetDefSpendKinds(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/spending/kind/") {
			handler.handleGetSpendKinds(w, r)
		} else {
			handler.handleUnknownPath(w)
		}
	case "POST":
		log.Error("not implemented")
	default:
		err := SendAPIErrorResp(w, "unknown request method", http.StatusBadRequest)
		if err != nil {
			log.Errorf("failed to send error response to client. unknown request method. details: %s", err.Error())
		}
	}
}

func (handler *SpendKindHandler) handleGetDefSpendKinds(w http.ResponseWriter, r *http.Request) {
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

func (handler *SpendKindHandler) handleUnknownPath(w http.ResponseWriter) {
	err := SendAPIErrorResp(w, "unknown path", http.StatusBadRequest)
	if err != nil {
		log.Errorf("error while sending error response to client [unknown path]: %s", err.Error())
	}
}
