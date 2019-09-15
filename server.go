package ispend

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func routerSetup() (r *mux.Router) {
	r = mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Error(err.Error())
		}
	})

	usersHandler := &UsersHandler{}
	spendingHandler := &SpendingHandler{}

	// new users, list users, etc ...
	r.Handle("/users", usersHandler)
	// new spending, remove spending, update spendings ...
	r.Handle("/spending/{username}", spendingHandler)

	return r
}

func Serve() {
	router := routerSetup()
	ipAndPort := fmt.Sprintf("%s:%s", IPAddress, Port)
	httpServer := &http.Server{
		Handler:      router,
		Addr:         ipAndPort,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Infof(" > server listening on: [%s]", ipAndPort)
	log.Fatal(httpServer.ListenAndServe())
}
