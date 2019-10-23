package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/2beens/ispend/internal/services"

	"github.com/2beens/ispend/internal/models"
	"github.com/2beens/ispend/internal/platform"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type SpendingHandler struct {
	usersService        *services.UsersService
	loginSessionManager *platform.LoginSessionManager
}

func SpendingHandlerSetup(router *mux.Router, usersService *services.UsersService, loginSessionManager *platform.LoginSessionManager) {
	handler := &SpendingHandler{
		usersService:        usersService,
		loginSessionManager: loginSessionManager,
	}

	router.HandleFunc("", handler.handleNewSpending).Methods("POST")
	router.HandleFunc("/{username}/{spendID}", handler.handleDeleteSpending).Methods("DELETE")
	router.HandleFunc("/id/{id}/{username}", handler.handleGetUserSpendingByID).Methods("GET")
	router.HandleFunc("/all/{username}", handler.handleGetUserSpends).Methods("GET")
}

func (handler *SpendingHandler) handleGetUserSpendingByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := r.Header.Get("X-Ispend-SessionID")
	if handler.loginSessionManager.IsUserNotLoggedIn(sessionID, username) {
		platform.SendAPIErrorResp(w, "must be logged in", http.StatusUnauthorized)
		return
	}

	spendID := vars["id"]
	user, err := handler.usersService.GetUser(username)
	if err != nil {
		platform.SendAPIErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i := range user.Spends {
		if user.Spends[i].ID == spendID {
			platform.SendAPIOKRespWithData(w, "success", user.Spends[i])
			return
		}
	}

	platform.SendAPIErrorResp(w, "not found", http.StatusNotFound)
}

func (handler *SpendingHandler) handleGetUserSpends(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	sessionID := r.Header.Get("X-Ispend-SessionID")
	if handler.loginSessionManager.IsUserNotLoggedIn(sessionID, username) {
		platform.SendAPIErrorResp(w, "must be logged in", http.StatusUnauthorized)
		return
	}

	user, err := handler.usersService.GetUser(username)
	if err != nil {
		platform.SendAPIErrorResp(w, err.Error(), http.StatusBadRequest)
		return
	}

	platform.SendAPIOKRespWithData(w, "success", user.Spends)
}

func (handler *SpendingHandler) handleDeleteSpending(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	if username == "" {
		platform.SendAPIErrorResp(w, "missing username", http.StatusBadRequest)
		return
	}

	sessionID := r.Header.Get("X-Ispend-SessionID")
	if handler.loginSessionManager.IsUserNotLoggedIn(sessionID, username) {
		platform.SendAPIErrorResp(w, "must be logged in", http.StatusUnauthorized)
		return
	}

	spendID := vars["spendID"]
	if spendID == "" {
		platform.SendAPIErrorResp(w, "missing spending ID", http.StatusBadRequest)
		return
	}

	err := handler.usersService.DeleteSpending(username, spendID)
	if err != nil {
		if err == platform.ErrNotFound {
			platform.SendAPIErrorResp(w, "not found", http.StatusNotFound)
		} else {
			platform.SendAPIErrorResp(w, "internal server error 93215", http.StatusInternalServerError)
		}
	} else {
		platform.SendAPIOKResp(w, "success")
	}
}

func (handler *SpendingHandler) handleNewSpending(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		return
	}

	username := r.FormValue("username")
	if username == "" {
		platform.SendAPIErrorResp(w, "missing username", http.StatusBadRequest)
		return
	}

	sessionID := r.Header.Get("X-Ispend-SessionID")
	if handler.loginSessionManager.IsUserNotLoggedIn(sessionID, username) {
		platform.SendAPIErrorResp(w, "must be logged in", http.StatusUnauthorized)
		return
	}

	currency := r.FormValue("currency")
	if currency == "" {
		platform.SendAPIErrorResp(w, "missing/wrong currency", http.StatusBadRequest)
		return
	}
	amountParam := r.FormValue("amount")
	amount, err := strconv.ParseFloat(amountParam, 32)
	if err != nil {
		log.Errorf("new spending, error 9004: %s", err.Error())
		platform.SendAPIErrorResp(w, "missing/wrong amount", http.StatusBadRequest)
		return
	}
	kindIdParam := r.FormValue("kind_id")
	kindId, _ := strconv.Atoi(kindIdParam)
	spendKind, err := handler.usersService.GetSpendKind(username, kindId)
	if err != nil {
		log.Errorf("new spending, error 9005: %s", err.Error())
		platform.SendAPIErrorResp(w, "missing/wrong spending kind ID", http.StatusBadRequest)
		return
	}

	user, err := handler.usersService.GetUser(username)
	if err != nil && err != platform.ErrNotFound {
		log.Errorf("new spending, error 9003: %s", err.Error())
		platform.SendAPIErrorResp(w, "server error 9003", http.StatusInternalServerError)
		return
	}
	if err == platform.ErrNotFound {
		platform.SendAPIErrorResp(w, "user not found", http.StatusBadRequest)
		return
	}

	spending := models.Spending{
		//ID:       GenerateRandomString(10),
		Currency: currency,
		Amount:   float32(amount),
		Kind:     spendKind,
		// more accurate would be to take the client timestamp
		Timestamp: time.Now(),
	}

	// will also add this spending to user.spends
	err = handler.usersService.StoreSpending(user, spending)
	if err != nil {
		log.Errorf("new spending, error 9004: %s", err.Error())
		platform.SendAPIErrorResp(w, "server error 9004", http.StatusInternalServerError)
		return
	}

	log.Tracef("new spending added: %v", spending)

	apiErr := models.APIResponse{Status: http.StatusOK, Message: "success", IsError: false, Data: spending.ID}
	platform.SendAPIResp(w, apiErr)
}
