package ispend

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type UsersHandler struct {
	db                  SpenderDB
	loginSessionHandler *LoginSessionHandler
}

func NewUsersHandler(db SpenderDB, loginSessionHandler *LoginSessionHandler) *UsersHandler {
	return &UsersHandler{
		db:                  db,
		loginSessionHandler: loginSessionHandler,
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
		} else if r.URL.Path == "/users/login" {
			handler.handleLogin(w, r)
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

func (handler *UsersHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		return
	}

	password := r.FormValue("password")
	if password == "" {
		_ = SendAPIErrorResp(w, "missing password", http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	if username == "" {
		_ = SendAPIErrorResp(w, "missing username", http.StatusBadRequest)
		return
	}

	user, err := handler.db.GetUser(username)
	if err != nil && err != ErrNotFound {
		log.Errorf("error while logging user: %s", err.Error())
		_ = SendAPIErrorResp(w, "server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		_ = SendAPIErrorResp(w, "error, user does not exists", http.StatusBadRequest)
		return
	}
	if user.Password != password {
		_ = SendAPIErrorResp(w, "wrong username/password", http.StatusBadRequest)
		return
	}

	cookieID := handler.loginSessionHandler.New(username)
	_ = SendAPIOKRespWithData(w, "success", cookieID)
}

func (handler *UsersHandler) handleNewUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		return
	}

	password := r.FormValue("password")
	if password == "" {
		_ = SendAPIErrorResp(w, "missing password", http.StatusBadRequest)
		return
	}

	// TODO: hash password

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

	log.Tracef("creating new user [%s], pass:[%s]...", username, password)

	spKinds, err := handler.db.GetAllDefaultSpendKinds()
	if err != nil {
		log.Errorf("error getting spend kinds: %s", err.Error())
		_ = SendAPIErrorResp(w, "server error", http.StatusInternalServerError)
		return
	}
	user := NewUser(username, password, spKinds)
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

	log.Tracef("creating new user [%s] created", username)
}

func (handler *UsersHandler) handleUnknownPath(w http.ResponseWriter) {
	err := SendAPIErrorResp(w, "unknown path", http.StatusBadRequest)
	if err != nil {
		log.Errorf("error while sending error response to client [unknown path]: %s", err.Error())
	}
}
