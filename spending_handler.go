package ispend

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type SpendingHandler struct {
	usersService        *UsersService
	loginSessionManager *LoginSessionManager
}

func SpendingHandlerSetup(router *mux.Router, usersService *UsersService, loginSessionManager *LoginSessionManager) {
	handler := &SpendingHandler{
		usersService:        usersService,
		loginSessionManager: loginSessionManager,
	}

	router.HandleFunc("", handler.handleNewSpending)
	router.HandleFunc("/id/{id}/{username}", handler.handleGetUserSpendingByID)
	router.HandleFunc("/all/{username}", handler.handleGetUserSpends)
}

func (handler *SpendingHandler) handleGetUserSpendingByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := r.Header.Get("X-Ispend-SessionID")
	if handler.loginSessionManager.IsUserNotLoggedIn(sessionID, username) {
		_ = SendAPIErrorResp(w, "must be logged in", http.StatusUnauthorized)
		return
	}

	spendID := vars["id"]
	user, err := handler.usersService.GetUser(username)
	if err != nil {
		sendErr := SendAPIErrorResp(w, err.Error(), http.StatusBadRequest)
		if sendErr != nil {
			log.Errorf("error while sending error response to client [get user spending]: %s", sendErr.Error())
		}
		return
	}

	for i := range user.Spends {
		if user.Spends[i].ID == spendID {
			err = SendAPIOKRespWithData(w, "success", user.Spends[i])
			if err != nil {
				log.Errorf("error while sending user spends response to client [get user spending]: %s", err.Error())
			}
			return
		}
	}

	_ = SendAPIErrorResp(w, "not found", http.StatusNotFound)
}

func (handler *SpendingHandler) handleGetUserSpends(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := r.Header.Get("X-Ispend-SessionID")
	if handler.loginSessionManager.IsUserNotLoggedIn(sessionID, username) {
		_ = SendAPIErrorResp(w, "must be logged in", http.StatusUnauthorized)
		return
	}

	user, err := handler.usersService.GetUser(username)
	if err != nil {
		sendErr := SendAPIErrorResp(w, err.Error(), http.StatusBadRequest)
		if sendErr != nil {
			log.Errorf("error while sending error response to client [get all user spends]: %s", sendErr.Error())
		}
		return
	}

	err = SendAPIOKRespWithData(w, "success", user.Spends)
	if err != nil {
		log.Errorf("error while sending user spends response to client [get all spends]: %s", err.Error())
	}
}

func (handler *SpendingHandler) handleNewSpending(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		return
	}

	username := r.FormValue("username")
	if username == "" {
		_ = SendAPIErrorResp(w, "missing username", http.StatusBadRequest)
		return
	}

	sessionID := r.Header.Get("X-Ispend-SessionID")
	if handler.loginSessionManager.IsUserNotLoggedIn(sessionID, username) {
		_ = SendAPIErrorResp(w, "must be logged in", http.StatusUnauthorized)
		return
	}

	currency := r.FormValue("currency")
	if currency == "" {
		_ = SendAPIErrorResp(w, "missing/wrong currency", http.StatusBadRequest)
		return
	}
	amountParam := r.FormValue("amount")
	amount, err := strconv.ParseFloat(amountParam, 32)
	if err != nil {
		log.Errorf("new spending, error 9004: %s", err.Error())
		_ = SendAPIErrorResp(w, "missing/wrong amount", http.StatusBadRequest)
		return
	}
	kindIdParam := r.FormValue("kind_id")
	kindId, _ := strconv.Atoi(kindIdParam)
	spendKind, err := handler.usersService.GetSpendKind(username, kindId)
	if err != nil {
		log.Errorf("new spending, error 9005: %s", err.Error())
		_ = SendAPIErrorResp(w, "missing/wrong spending kind ID", http.StatusBadRequest)
		return
	}

	user, err := handler.usersService.GetUser(username)
	if err != nil && err != ErrNotFound {
		log.Errorf("new spending, error 9003: %s", err.Error())
		_ = SendAPIErrorResp(w, "server error 9003", http.StatusInternalServerError)
		return
	}
	if err == ErrNotFound {
		_ = SendAPIErrorResp(w, "user not found", http.StatusBadRequest)
		return
	}

	spending := Spending{
		//ID:       GenerateRandomString(10),
		Currency: currency,
		Amount:   float32(amount),
		Kind:     spendKind,
		// more accurate would be to take the client timestamp
		Timestamp: time.Now(),
	}

	// will also add this spending to user.spends
	id, err := handler.usersService.StoreSpending(username, spending)
	if err != nil {
		log.Errorf("new spending, error 9004: %s", err.Error())
		_ = SendAPIErrorResp(w, "server error 9004", http.StatusInternalServerError)
		return
	}

	// TODO: check is this needed; how user is taken currently
	spending.ID = id
	user.Spends = append(user.Spends, spending)

	log.Tracef("new spending added: %v", spending)

	apiErr := APIResponse{Status: http.StatusOK, Message: "success", IsError: false, Data: spending.ID}
	_ = SendAPIResp(w, apiErr)
}
