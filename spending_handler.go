package ispend

import (
	"net/http"

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
		log.Error("not implemented")
	default:
		err := SendAPIErrorResp(w, "unknown request method", http.StatusBadRequest)
		if err != nil {
			log.Errorf("failed to send error response to client. unknown request method. details: %s", err.Error())
		}
	}
}
