package ispend

import (
	"github.com/gorilla/mux"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type SpendingHandler struct {
	db SpenderDB
}

func NewSpendingHandler(db SpenderDB) *SpendingHandler {
	return &SpendingHandler{
		db: db,
	}
}

func (handler *SpendingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		log.Error("not implemented")
	case "POST":
		if strings.HasPrefix(r.URL.Path, "") {
			handler.handleNewSpending(w, r)
		} else {
			handler.handleUnknownPath(w)
		}
	default:
		err := SendAPIErrorResp(w, "unknown request method", http.StatusBadRequest)
		if err != nil {
			log.Errorf("failed to send error response to client. unknown request method. details: %s", err.Error())
		}
	}
}

func (handler *SpendingHandler) handleNewSpending(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	if username == "" {
		_ = SendAPIErrorResp(w, "wrong username", http.StatusBadRequest)
		return
	}

	user, err := handler.db.GetUser(username)
	if err != nil {
		_ = SendAPIErrorResp(w, "server error 9002", http.StatusInternalServerError)
		log.Warnf("error [%s]: %s", r.URL.Path, err.Error())
		return
	}

	_ = user
}

func (handler *SpendingHandler) handleUnknownPath(w http.ResponseWriter) {
	err := SendAPIErrorResp(w, "unknown path", http.StatusBadRequest)
	if err != nil {
		log.Errorf("error while sending error response to client [unknown path]: %s", err.Error())
	}
}
