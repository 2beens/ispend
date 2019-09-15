package ispend

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type UsersHandler struct {
	db SpenderDB
}

func NewUsersHandler(db SpenderDB) *UsersHandler {
	return &UsersHandler{
		db: db,
	}
}

func (handler *UsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if r.URL.Path == "/users" {
			handler.handleGetAllUsers(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/users/") {
			handler.handleGetUser(w, r)
		} else {
			err := SendAPIErrorResp(w, "unknown path", http.StatusBadRequest)
			if err != nil {
				log.Errorf("error while sending error response to client [unknown path]: %s", err.Error())
			}
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

func (handler *UsersHandler) handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	users := handler.db.GetAllUsers()
	err := SendAPIOKRespWithData(w, "success", users)
	if err != nil {
		log.Errorf("error while sending response to client [get all users]: %s", err.Error())
	}
}

func (handler *UsersHandler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	user, err := handler.db.GetUser(username)
	if err != nil {
		sendErr := SendAPIErrorResp(w, err.Error(), http.StatusBadRequest)
		if sendErr != nil {
			log.Errorf("error while sending error response to client [get user]: %s", sendErr.Error())
		}
		return
	}
	err = SendAPIOKRespWithData(w, "success", user)
	if err != nil {
		log.Errorf("error while sending response to client [get user]: %s", err.Error())
	}
}
