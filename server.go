package ispend

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Tracef(" ====> request path: [%s]", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func routerSetup(db SpenderDB) (r *mux.Router) {
	r = mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Error(err.Error())
		}
	})

	viewsMaker, err := NewViewsMaker("public/views/")
	if err != nil {
		log.Fatal(err.Error())
	}
	r.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		viewsMaker.RenderView(w, "index", nil)
	})

	usersHandler := NewUsersHandler(db)
	spendingHandler := NewSpendingHandler(db)
	spendKindHandler := NewSpendKindHandler(db)

	// new users, list users, etc ...
	r.Handle("/users", usersHandler)
	r.Handle("/users/{username}", usersHandler)
	// new spending, remove spending, update spendings ...
	r.Handle("/spending/{username}", spendingHandler)
	// new spend kind, spend kinds list, etc ...
	r.Handle("/spending/kind", spendKindHandler)
	r.Handle("/spending/kind/{username}", spendKindHandler)

	r.Use(loggingMiddleware)

	return r
}

func Serve(port string) {
	// TODO: will be adapted ...
	TestPostgresDB()

	tempDB := prepareTempDB()
	for _, u := range tempDB.Users {
		log.Debugf("user: %s", u.Username)
	}

	if port == "" {
		port = DefaultPort
		log.Debugf("using default port: %s", port)
	}

	router := routerSetup(tempDB)
	ipAndPort := fmt.Sprintf("%s:%s", IPAddress, port)
	httpServer := &http.Server{
		Handler:      router,
		Addr:         ipAndPort,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Infof(" > server listening on: [%s]", ipAndPort)
	log.Fatal(httpServer.ListenAndServe())
}

func prepareTempDB() *TempDB {
	skNightlife := SpendKind{"nightlife"}
	skTravel := SpendKind{"travel"}
	skFood := SpendKind{"food"}
	skRent := SpendKind{"rent"}
	defSpendKinds := []SpendKind{skNightlife, skTravel, skFood, skRent}

	adminUser := NewUser("admin", defSpendKinds)
	adminUser.Spendings = append(adminUser.Spendings, Spending{
		Amount:   100,
		Currency: "RSD",
		Kind:     skNightlife,
	})
	adminUser.Spendings = append(adminUser.Spendings, Spending{
		Amount:   2300,
		Currency: "RSD",
		Kind:     skTravel,
	})
	lazarUser := NewUser("lazar", defSpendKinds)
	lazarUser.Spendings = append(lazarUser.Spendings, Spending{
		Amount:   89.99,
		Currency: "USD",
		Kind:     skTravel,
	})

	tempDB := NewTempDB()
	err := tempDB.StoreUser(adminUser)
	if err != nil {
		log.Panic(err.Error())
	}
	err = tempDB.StoreUser(lazarUser)
	if err != nil {
		log.Panic(err.Error())
	}

	err = tempDB.StoreDefaultSpendKind(skNightlife)
	if err != nil {
		log.Panic(err.Error())
	}
	err = tempDB.StoreDefaultSpendKind(skFood)
	if err != nil {
		log.Panic(err.Error())
	}
	err = tempDB.StoreDefaultSpendKind(skRent)
	if err != nil {
		log.Panic(err.Error())
	}
	err = tempDB.StoreDefaultSpendKind(skTravel)
	if err != nil {
		log.Panic(err.Error())
	}

	return tempDB
}
