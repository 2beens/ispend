package internal

import (
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type UsersHandler struct {
	router              *mux.Router
	usersService        *UsersService
	loginSessionManager *LoginSessionManager
}

func UsersHandlerSetup(router *mux.Router, usersService *UsersService, loginSessionManager *LoginSessionManager) {
	handler := &UsersHandler{
		router:              router,
		usersService:        usersService,
		loginSessionManager: loginSessionManager,
	}

	router.HandleFunc("", handler.handleGetAllUsers).Methods("GET")
	router.HandleFunc("", handler.handleNewUser).Methods("POST")
	router.HandleFunc("/me/{username}/{cookie}", handler.handleGetMe)
	router.HandleFunc("/login", handler.handleLogin)
	router.HandleFunc("/login/check", handler.handleCheckSessionID)
	router.HandleFunc("/logout", handler.handleLogout)
	router.HandleFunc("/{username}", handler.handleGetUser)
}

func (handler *UsersHandler) handleGetMe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	// TODO: cookie should be transported via body/header, not query param
	cookie := vars["cookie"]
	if username == "" {
		SendAPIErrorResp(w, "wrong username", http.StatusBadRequest)
		return
	}
	if cookie == "" {
		SendAPIErrorResp(w, "must be logged in", http.StatusBadRequest)
		return
	}

	loginSession, err := handler.loginSessionManager.GetBySessionID(cookie)
	if err != nil {
		SendAPIErrorResp(w, "server error 9001", http.StatusInternalServerError)
		log.Warnf("error [%s]: %s", r.URL.Path, err.Error())
		return
	}

	user, err := handler.usersService.GetUser(loginSession.Username)
	if err != nil {
		SendAPIErrorResp(w, "server error 9002", http.StatusInternalServerError)
		log.Warnf("error [%s]: %s", r.URL.Path, err.Error())
		return
	}

	userDto := NewUserDTO(user)
	SendAPIOKRespWithData(w, "success", userDto)
}

func (handler *UsersHandler) handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		return
	}

	username := r.FormValue("username")
	sessionID := r.Header.Get("X-Ispend-SessionID")
	if handler.loginSessionManager.IsUserNotLoggedIn(sessionID, username) {
		SendAPIErrorResp(w, "must be logged in", http.StatusUnauthorized)
		return
	}

	users, err := handler.usersService.GetAllUsers()
	if err != nil {
		SendAPIErrorResp(w, "internal server error 10002", http.StatusInternalServerError)
		log.Warnf("error getting all users [handleGetAllUsers]: %s", err.Error())
		return
	}
	SendAPIOKRespWithData(w, "success", users)
}

func (handler *UsersHandler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	user, err := handler.usersService.GetUser(username)
	if err != nil {
		SendAPIErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}
	SendAPIOKRespWithData(w, "success", user)
}

func (handler *UsersHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		return
	}

	cookieId := r.FormValue("sessionId")
	if cookieId == "" {
		SendAPIErrorResp(w, "missing sessionId", http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	if username == "" {
		SendAPIErrorResp(w, "missing username", http.StatusBadRequest)
		return
	}

	log.Tracef(" > logout user: [%s][%s]", username, cookieId)

	session, err := handler.loginSessionManager.GetBySessionID(cookieId)
	if err != nil {
		if err == ErrNotFound {
			SendAPIErrorResp(w, "error, session not found", http.StatusNotFound)
		} else {
			log.Errorf("logout error: %s", err.Error())
			SendAPIErrorResp(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	if session.Username != username {
		log.Errorf("error 10102, s. username [%s], username: %s", session.Username, username)
		SendAPIErrorResp(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = handler.loginSessionManager.Remove(session.Username)
	if err != nil {
		if err == ErrNotFound {
			log.Errorf("error 10103, s. username [%s], username: %s", session.Username, username)
			SendAPIErrorResp(w, "error, session not found", http.StatusNotFound)
		} else {
			log.Errorf("logout error: %s", err.Error())
			SendAPIErrorResp(w, "internal server error", http.StatusInternalServerError)
		}
	}

	SendAPIOKResp(w, "success")
}

func (handler *UsersHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		return
	}

	password := r.FormValue("password")
	if password == "" {
		SendAPIErrorResp(w, "missing password", http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	if username == "" {
		SendAPIErrorResp(w, "missing username", http.StatusBadRequest)
		return
	}

	user, err := handler.usersService.GetUser(username)
	if err != nil && err != ErrNotFound {
		log.Errorf("error while logging user: %s", err.Error())
		SendAPIErrorResp(w, "server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		SendAPIErrorResp(w, "error, user does not exists", http.StatusBadRequest)
		return
	}
	if !CheckPasswordHash(password, user.Password) {
		SendAPIErrorResp(w, "wrong username/password", http.StatusBadRequest)
		return
	}

	session, err := handler.loginSessionManager.GetByUsername(username)
	if err == nil && session != nil {
		SendAPIOKRespWithData(w, "success", session.SessionID)
		return
	}

	cookieID := handler.loginSessionManager.New(username)
	SendAPIOKRespWithData(w, "success", cookieID)
}

func (handler *UsersHandler) handleNewUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		SendAPIErrorResp(w, "internal server error 109011", http.StatusInternalServerError)
		return
	}

	password := r.FormValue("password")
	if password == "" {
		SendAPIErrorResp(w, "missing password", http.StatusBadRequest)
		return
	}

	// hash password
	passwordHash, err := HashPassword(password)
	if err != nil {
		log.Errorf("error [getting password hash] while adding new user: %s", err)
		SendAPIErrorResp(w, "server error 111111", http.StatusInternalServerError)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		SendAPIErrorResp(w, "missing username", http.StatusBadRequest)
		return
	}
	existingUser, err := handler.usersService.GetUser(username)
	if err != nil && err != ErrNotFound {
		log.Errorf("error while adding new user: %s", err.Error())
		SendAPIErrorResp(w, "server error", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		SendAPIErrorResp(w, "error, user exists", http.StatusConflict)
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
		SendAPIErrorResp(w, "server error", http.StatusInternalServerError)
		return
	}
	user := NewUser(email, username, passwordHash, spKinds)
	err = handler.usersService.AddUser(user)
	if err != nil {
		log.Errorf("error while adding new user: %s", err.Error())
		SendAPIErrorResp(w, "server error", http.StatusInternalServerError)
		return
	}

	SendAPIOKResp(w, "success")
	log.Tracef("new user [%s] created", username)
}

func (handler *UsersHandler) handleCheckSessionID(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		return
	}

	sessionID := r.FormValue("sessionId")
	if sessionID == "" {
		SendAPIErrorResp(w, "missing sessionID", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		SendAPIErrorResp(w, "missing username", http.StatusBadRequest)
		return
	}

	session, err := handler.loginSessionManager.GetBySessionID(sessionID)
	if err != nil && err != ErrNotFound {
		log.Errorf("check session id error: %s", err)
		SendAPIErrorResp(w, "internal server error 109013", http.StatusInternalServerError)
		return
	}
	if err == ErrNotFound || (session != nil && session.Username != username) {
		SendAPIOKResp(w, "false")
		return
	}

	SendAPIOKResp(w, "true")
}
