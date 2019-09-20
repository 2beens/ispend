package ispend

import (
	"context"
	"fmt"
	"net/http"
	"os"
	ossignal "os/signal"
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

func routerSetup(db SpenderDB, chInterrupt chan signal) (r *mux.Router) {
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
	r.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		viewsMaker.RenderView(w, "debug", nil)
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

	usersHandler := NewUsersHandler(db, loginSessionManager)
	spendingHandler := NewSpendingHandler(db)
	spendKindHandler := NewSpendKindHandler(db)

	// new users, list users, etc ...
	r.Handle("/users", usersHandler)
	r.Handle("/users/login", usersHandler)
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
	// TODO: make a type out of this file ?

	// TODO: will be adapted ...
	TestPostgresDB()

	chInterrupt := make(chan signal, 1)
	chOsInterrupt := make(chan os.Signal, 1)
	ossignal.Notify(chOsInterrupt, os.Interrupt)

	tempDB := NewTempDB()
	for _, u := range tempDB.Users {
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

	router := routerSetup(tempDB, chInterrupt)
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
	gracefulShutdown(httpServer)
}

func gracefulShutdown(httpServer *http.Server) {
	log.Debug("graceful shutdown initiated ...")

	maxWaitDuration := time.Second * 15
	ctx, cancel := context.WithTimeout(context.Background(), maxWaitDuration)
	defer cancel()
	err := httpServer.Shutdown(ctx)
	if err != nil {
		log.Error(" >>> failed to gracefully shutdown")
	}

	log.Warn("server shut down")
	os.Exit(0)
}
