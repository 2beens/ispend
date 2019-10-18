package ispend

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type UsersHandler struct {
	usersService        *UsersService
	loginSessionManager *LoginSessionManager
}

func NewUsersHandler(usersService *UsersService, loginSessionManager *LoginSessionManager) *UsersHandler {
	return &UsersHandler{
		usersService:        usersService,
		loginSessionManager: loginSessionManager,
	}
}

func (handler *UsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if r.URL.Path == "/users" {
			handler.handleGetAllUsers(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/users/me/") {
			handler.handleGetMe(w, r)
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
		} else if r.URL.Path == "/users/login/check" {
			handler.checkSessionID(w, r)
		} else if r.URL.Path == "/users/logout" {
			handler.handleLogout(w, r)
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

func (handler *UsersHandler) handleGetMe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	// TODO: cookie should be transported via body/header, not query param
	cookie := vars["cookie"]
	if username == "" {
		_ = SendAPIErrorResp(w, "wrong username", http.StatusBadRequest)
		return
	}
	if cookie == "" {
		_ = SendAPIErrorResp(w, "must be logged in", http.StatusBadRequest)
		return
	}

	loginSession, err := handler.loginSessionManager.GetBySessionID(cookie)
	if err != nil {
		_ = SendAPIErrorResp(w, "server error 9001", http.StatusInternalServerError)
		log.Warnf("error [%s]: %s", r.URL.Path, err.Error())
		return
	}

	user, err := handler.usersService.GetUser(loginSession.Username)
	if err != nil {
		_ = SendAPIErrorResp(w, "server error 9002", http.StatusInternalServerError)
		log.Warnf("error [%s]: %s", r.URL.Path, err.Error())
		return
	}

	userDto := NewUserDTO(user)
	_ = SendAPIOKRespWithData(w, "success", userDto)
}

func (handler *UsersHandler) handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := handler.usersService.GetAllUsers()
	if err != nil {
		_ = SendAPIErrorResp(w, "internal server error 10002", http.StatusInternalServerError)
		log.Warnf("error getting all users [handleGetAllUsers]: %s", err.Error())
		return
	}
	err = SendAPIOKRespWithData(w, "success", users)
	if err != nil {
		log.Errorf("error while sending response to client [get all users]: %s", err.Error())
	}
}

func (handler *UsersHandler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	user, err := handler.usersService.GetUser(username)
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

func (handler *UsersHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		return
	}

	cookieId := r.FormValue("sessionId")
	if cookieId == "" {
		_ = SendAPIErrorResp(w, "missing sessionId", http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	if username == "" {
		_ = SendAPIErrorResp(w, "missing username", http.StatusBadRequest)
		return
	}

	log.Tracef(" > logout user: [%s][%s]", username, cookieId)

	session, err := handler.loginSessionManager.GetBySessionID(cookieId)
	if err != nil {
		if err == ErrNotFound {
			_ = SendAPIErrorResp(w, "error, session not found", http.StatusNotFound)
		} else {
			log.Errorf("logout error: %s", err.Error())
			_ = SendAPIErrorResp(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	if session.Username != username {
		log.Errorf("error 10102, s. username [%s], username: %s", session.Username, username)
		_ = SendAPIErrorResp(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = handler.loginSessionManager.Remove(session.Username)
	if err != nil {
		if err == ErrNotFound {
			log.Errorf("error 10103, s. username [%s], username: %s", session.Username, username)
			_ = SendAPIErrorResp(w, "error, session not found", http.StatusNotFound)
		} else {
			log.Errorf("logout error: %s", err.Error())
			_ = SendAPIErrorResp(w, "internal server error", http.StatusInternalServerError)
		}
	}

	_ = SendAPIOKResp(w, "success")
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

	user, err := handler.usersService.GetUser(username)
	if err != nil && err != ErrNotFound {
		log.Errorf("error while logging user: %s", err.Error())
		_ = SendAPIErrorResp(w, "server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		_ = SendAPIErrorResp(w, "error, user does not exists", http.StatusBadRequest)
		return
	}
	if !CheckPasswordHash(password, user.Password) {
		_ = SendAPIErrorResp(w, "wrong username/password", http.StatusBadRequest)
		return
	}

	session, err := handler.loginSessionManager.GetByUsername(username)
	if err == nil && session != nil {
		_ = SendAPIOKRespWithData(w, "success", session.SessionID)
		return
	}

	cookieID := handler.loginSessionManager.New(username)
	_ = SendAPIOKRespWithData(w, "success", cookieID)
}

func (handler *UsersHandler) handleNewUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		_ = SendAPIErrorResp(w, "internal server error 109011", http.StatusInternalServerError)
		return
	}

	password := r.FormValue("password")
	if password == "" {
		_ = SendAPIErrorResp(w, "missing password", http.StatusBadRequest)
		return
	}

	// hash password
	passwordHash, err := HashPassword(password)
	if err != nil {
		log.Errorf("error [getting password hash] while adding new user: %s", err)
		_ = SendAPIErrorResp(w, "server error 111111", http.StatusInternalServerError)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		_ = SendAPIErrorResp(w, "missing username", http.StatusBadRequest)
		return
	}
	existingUser, err := handler.usersService.GetUser(username)
	if err != nil && err != ErrNotFound {
		log.Errorf("error while adding new user: %s", err.Error())
		_ = SendAPIErrorResp(w, "server error", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		_ = SendAPIErrorResp(w, "error, user exists", http.StatusConflict)
		return
	}

	email := r.FormValue("email")
	if email == "" {
		log.Tracef("new user - missing email")
	}

	log.Tracef("creating new user [%s], pass [%s] ...", username, passwordHash)

	spKinds, err := handler.usersService.GetAllDefaultSpendKinds()
	if err != nil {
		log.Errorf("error getting spend kinds: %s", err.Error())
		_ = SendAPIErrorResp(w, "server error", http.StatusInternalServerError)
		return
	}
	user := NewUser(email, username, passwordHash, spKinds)
	err = handler.usersService.AddUser(user)
	if err != nil {
		log.Errorf("error while adding new user: %s", err.Error())
		_ = SendAPIErrorResp(w, "server error", http.StatusInternalServerError)
		return
	}

	err = SendAPIOKResp(w, "success")
	if err != nil {
		log.Errorf("error while adding new user: %s", err.Error())
	}

	log.Tracef("new user [%s] created", username)
}

func (handler *UsersHandler) handleUnknownPath(w http.ResponseWriter) {
	err := SendAPIErrorResp(w, "unknown path", http.StatusBadRequest)
	if err != nil {
		log.Errorf("error while sending error response to client [unknown path]: %s", err.Error())
	}
}

func (handler *UsersHandler) checkSessionID(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		return
	}

	sessionID := r.FormValue("sessionId")
	if sessionID == "" {
		_ = SendAPIErrorResp(w, "missing sessionID", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		_ = SendAPIErrorResp(w, "missing username", http.StatusBadRequest)
		return
	}

	session, err := handler.loginSessionManager.GetBySessionID(sessionID)
	if err != nil && err != ErrNotFound {
		log.Errorf("check session id error: %s", err)
		_ = SendAPIErrorResp(w, "internal server error 109013", http.StatusInternalServerError)
		return
	}

	if err == ErrNotFound || (session != nil && session.Username != username) {
		_ = SendAPIOKResp(w, "false")
		return
	}

	_ = SendAPIOKResp(w, "true")
}
