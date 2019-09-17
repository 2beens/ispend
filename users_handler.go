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
			handler.handleUnknownPath(w)
		}
	case "POST":
		if r.URL.Path == "/users" {
			handler.handleNewUser(w, r)
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

func (handler *UsersHandler) handleNewUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [/users]: %s", err.Error())
		return
	}

	username := r.FormValue("username")
	if username == "" {
		_ = SendAPIErrorResp(w, "missing username", http.StatusBadRequest)
		return
	}
	existingUser, err := handler.db.GetUser(username)
	if err != nil && err != ErrNotFound {
		log.Errorf("error while adding new user: %s", err.Error())
		_ = SendAPIErrorResp(w, "server error", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		_ = SendAPIErrorResp(w, "error, user exists", http.StatusConflict)
		return
	}

	spKinds, err := handler.db.GetAllDefaultSpendKinds()
	if err != nil {
		log.Errorf("error getting spend kinds: %s", err.Error())
		_ = SendAPIErrorResp(w, "server error", http.StatusInternalServerError)
		return
	}
	user := NewUser(username, spKinds)
	err = handler.db.StoreUser(user)
	if err != nil {
		log.Errorf("error while adding new user: %s", err.Error())
		_ = SendAPIErrorResp(w, "server error", http.StatusInternalServerError)
		return
	}

	err = SendAPIOKResp(w, "success")
	if err != nil {
		log.Errorf("error while adding new user: %s", err.Error())
	}
}

func (handler *UsersHandler) handleUnknownPath(w http.ResponseWriter) {
	err := SendAPIErrorResp(w, "unknown path", http.StatusBadRequest)
	if err != nil {
		log.Errorf("error while sending error response to client [unknown path]: %s", err.Error())
	}
}
