package ispend

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type SpendingHandler struct {
	usersService        *UsersService
	loginSessionManager *LoginSessionManager
}

func NewSpendingHandler(usersService *UsersService, loginSessionManager *LoginSessionManager) *SpendingHandler {
	return &SpendingHandler{
		usersService:        usersService,
		loginSessionManager: loginSessionManager,
	}
}

func (handler *SpendingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if strings.HasPrefix(r.URL.Path, "/spending/id/") {
			handler.handleGetUserSpendingByID(w, r)
		} else if strings.HasPrefix(r.URL.Path, "/spending/all/") {
			handler.handleGetUserSpends(w, r)
		} else {
			handler.handleUnknownPath(w)
		}
	case "POST":
		if r.URL.Path == "/spending" {
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

	for i, _ := range user.Spends {
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

func (handler *SpendingHandler) handleUnknownPath(w http.ResponseWriter) {
	err := SendAPIErrorResp(w, "unknown path", http.StatusBadRequest)
	if err != nil {
		log.Errorf("error while sending error response to client [unknown path]: %s", err.Error())
	}
}
