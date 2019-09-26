package ispend

import (
	"context"
	"fmt"
	"net/http"
	"os"
	ossignal "os/signal"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var loginSessionManager = NewLoginSessionHandler()

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Tracef(" ====> request path: [%s]", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func panicRecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func(reqPath string) {
			if r := recover(); r != nil {
				log.Errorf(" >>> recovering from panic [path: %s]. error details: %v", reqPath, r)
				log.Error(" >>> stack trace: ")
				log.Error(string(debug.Stack()))
			}
		}(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func routerSetup(isProduction bool, db SpenderDB, chInterrupt chan signal) (r *mux.Router) {
	r = mux.NewRouter()

	// server static files
	fs := http.FileServer(http.Dir("./public/"))
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", fs))

	viewsMaker, err := NewViewsMaker("public/views/")
	if err != nil {
		log.Fatal(err.Error())
	}

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		viewsMaker.RenderIndex(w)
	})
	r.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		viewsMaker.RenderIndex(w)
	})
	r.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		viewsMaker.RenderView(w, "contact", nil)
	})
	r.HandleFunc("/examples", func(w http.ResponseWriter, r *http.Request) {
		viewsMaker.RenderView(w, "examples", nil)
	})
	r.HandleFunc("/page", func(w http.ResponseWriter, r *http.Request) {
		viewsMaker.RenderView(w, "page", nil)
	})
	r.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		viewsMaker.RenderView(w, "register", nil)
	})

	r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		err := SendAPIOKResp(w, "Oh yeah...")
		if err != nil {
			log.Error(err.Error())
		}
	})

	r.HandleFunc("/harakiri", func(w http.ResponseWriter, r *http.Request) {
		chInterrupt <- emptySignal
		err := SendAPIOKResp(w, "Goodbye cruel world...")
		if err != nil {
			log.Error(err.Error())
		}
	})

	logsPath := "/Users/2beens/Documents/projects/ispend"
	if isProduction {
		logsPath = "/root/ispend"
	}

	usersHandler := NewUsersHandler(db, loginSessionManager)
	spendingHandler := NewSpendingHandler(db, loginSessionManager)
	spendKindHandler := NewSpendKindHandler(db, loginSessionManager)
	debugHandler := NewDebugHandler(viewsMaker, logsPath, "ispend.log")

	// new users, list users, etc ...
	r.Handle("/users", usersHandler)
	r.Handle("/users/me/{username}/{cookie}", usersHandler)
	r.Handle("/users/login", usersHandler)
	r.Handle("/users/{username}", usersHandler)

	// new spending, remove spending, update spendings ...
	r.Handle("/spending", spendingHandler)
	r.Handle("/spending/id/{id}/{username}", spendingHandler)
	r.Handle("/spending/all/{username}", spendingHandler)

	// new spend kind, spend kinds list, etc ...
	r.Handle("/spending/kind", spendKindHandler)
	r.Handle("/spending/kind/{username}", spendKindHandler)

	// debug & misc
	r.Handle("/debug", debugHandler)
	r.Handle("/debug/logs", debugHandler)

	r.Use(loggingMiddleware)
	r.Use(panicRecoverMiddleware)

	return r
}

// TODO: make a type/struct out of this file ?
func Serve(port, environment, dbType string) {
	postgresDB := NewPostgresDBClient("localhost", 5432, "ispenddb", "2beens", "", "disable")
	err := postgresDB.Open()
	if err != nil {
		log.Errorf("cannot open PS DB connection: %s", err.Error())
	} else {
		allUsers, err := postgresDB.GetAllUsers()
		if err != nil {
			log.Error(err)
		}
		log.Debugf("gotten [%d] users from DB", len(allUsers))
		adminUser, err := postgresDB.GetUser("admin")
		if err != nil {
			log.Error(err)
		} else {
			log.Debugf("gotten user from DB: %s", adminUser.Email)
		}
		spendKinds, err := postgresDB.GetAllDefaultSpendKinds()
		log.Debugf("gotten [%d] def spend kinds", len(spendKinds))

		testSpend := Spending{
			Currency: "EUR",
			Amount:   100,
			Kind: &SpendKind{
				ID:   1,
				Name: "Sex",
			},
			Timestamp: time.Now(),
		}
		err = postgresDB.StoreSpending("admin", testSpend)
		if err != nil {
			log.Error(err)
		}
	}

	chInterrupt := make(chan signal, 1)
	chOsInterrupt := make(chan os.Signal, 1)
	ossignal.Notify(chOsInterrupt, os.Interrupt)

	inMemoryDB := NewInMemoryDB()
	for _, u := range inMemoryDB.Users {
		log.Debugf("user: %s", u.Username)
	}

	if port == "" {
		port = DefaultPort
		log.Debugf("using default port: %s", port)
	}

	// we need a webserver to get the pprof webserver
	pprofhost := "localhost"
	pprofport := "5002"
	go func() {
		log.Debugf("starting pprof server on [%s:%s] ...", pprofhost, pprofport)
		log.Debugln(http.ListenAndServe(pprofhost+":"+pprofport, nil))
	}()

	isProduction := environment == "p" || environment == "production"

	var router *mux.Router
	if strings.ToLower(dbType) == "mem" {
		router = routerSetup(isProduction, inMemoryDB, chInterrupt)
		log.Println(" > db: using in memory db")
	} else {
		router = routerSetup(isProduction, postgresDB, chInterrupt)
		log.Println(" > db: using Postgres db")
	}

	ipAndPort := fmt.Sprintf("%s:%s", IPAddress, port)

	httpServer := &http.Server{
		Handler:      router,
		Addr:         ipAndPort,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		log.Infof(" > server listening on: [%s]", ipAndPort)
		log.Fatal(httpServer.ListenAndServe())
	}()

	select {
	case <-chInterrupt:
		log.Warn("received harakiri signal, killing myself ...")
	case <-chOsInterrupt:
		log.Warn("os interrupt received ...")
	}
	gracefulShutdown(httpServer, postgresDB)
}

func gracefulShutdown(httpServer *http.Server, postgresDB *PostgresDBClient) {
	log.Debug("graceful shutdown initiated ...")

	err := postgresDB.Close()
	if err != nil {
		log.Warnf("failed to close postgres DB: " + err.Error())
	} else {
		log.Debug("postgres DB connection closed ...")
	}

	maxWaitDuration := time.Second * 15
	ctx, cancel := context.WithTimeout(context.Background(), maxWaitDuration)
	defer cancel()
	err = httpServer.Shutdown(ctx)
	if err != nil {
		log.Error(" >>> failed to gracefully shutdown")
	}

	log.Warn("server shut down")
	os.Exit(0)
}
