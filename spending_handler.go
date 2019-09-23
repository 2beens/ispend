package ispend

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SpendingHandler struct {
	db                  SpenderDB
	loginSessionManager *LoginSessionManager
}

func NewSpendingHandler(db SpenderDB, loginSessionManager *LoginSessionManager) *SpendingHandler {
	return &SpendingHandler{
		db:                  db,
		loginSessionManager: loginSessionManager,
	}
}

func (handler *SpendingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if strings.HasPrefix(r.URL.Path, "/spending/") {
			log.Error("not implemented")
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

func (handler *SpendingHandler) handleNewSpending(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Errorf("error parsing form values [%s]: %s", r.URL.Path, err.Error())
		return
	}

	// TODO: take cookie and see if user is logged, and if not - throw unauthorized error ??
	// 			still in development - will add later
	//cookie := r.FormValue("cookie")
	//username, err := handler.loginSessionManager.GetByCookieID(cookie)
	//if err != nil {
	//	_ = SendAPIErrorResp(w, "must be logged in", http.StatusUnauthorized)
	//	return
	//}

	username := r.FormValue("username")
	if username == "" {
		_ = SendAPIErrorResp(w, "missing username", http.StatusBadRequest)
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
	kindId := r.FormValue("kind_id")
	spendKind, err := handler.db.GetSpendKind(username, kindId)
	if err != nil {
		log.Errorf("new spending, error 9005: %s", err.Error())
		_ = SendAPIErrorResp(w, "missing/wrong spending kind ID", http.StatusBadRequest)
		return
	}

	user, err := handler.db.GetUser(username)
	if err != nil && err != ErrNotFound {
		log.Errorf("new spending, error 9003: %s", err.Error())
		_ = SendAPIErrorResp(w, "server error 9003", http.StatusInternalServerError)
		return
	}
	if err == ErrNotFound {
		_ = SendAPIErrorResp(w, "user not found", http.StatusBadRequest)
		return
	}

	_ = user

	spending := Spending{
		ID:       GenerateRandomString(10),
		Currency: currency,
		Amount:   float32(amount),
		Kind:     spendKind,
		// more accurate would be to take the client timestamp
		Timestamp: time.Now(),
	}

	// will also add this spending to user.spends
	err = handler.db.StoreSpending(username, spending)
	if err != nil {
		log.Errorf("new spending, error 9004: %s", err.Error())
		_ = SendAPIErrorResp(w, "server error 9004", http.StatusInternalServerError)
		return
	}

	_ = SendAPIOKResp(w, "success")
}

func (handler *SpendingHandler) handleUnknownPath(w http.ResponseWriter) {
	err := SendAPIErrorResp(w, "unknown path", http.StatusBadRequest)
	if err != nil {
		log.Errorf("error while sending error response to client [unknown path]: %s", err.Error())
	}
}
