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

var dbClient SpenderDB

func getLoggingMiddleware(graphiteClient *GraphiteClient) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: mute path logs for now
			sessionID := r.Header.Get("X-Ispend-SessionID")
			log.Tracef(" ====> request path: [%s] [sessionID: %s]", r.URL.Path, sessionID)
			path := r.URL.Path
			if path == "/" {
				path = "<root>"
			} else if strings.HasPrefix(path, "/") {
				path = path[1:]
			}
			graphiteClient.SimpleSendInt("paths."+path, 1)
			next.ServeHTTP(w, r)
		})
	}
}

func getPanicRecoverMiddleware(graphiteClient *GraphiteClient) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func(reqPath string) {
				if r := recover(); r != nil {
					log.Errorf(" >>> recovering from panic [path: %s]. error details: %v", reqPath, r)
					log.Error(" >>> stack trace: ")
					log.Error(string(debug.Stack()))

					path := reqPath
					if path == "/" {
						path = "<root>"
					} else if strings.HasPrefix(path, "/") {
						path = path[1:]
					}
					graphiteClient.SimpleSendInt("panic.recovery."+path, 1)
				}
			}(r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

func routerSetup(isProduction bool, db SpenderDB, graphiteClient *GraphiteClient, chInterrupt chan signal) (r *mux.Router) {
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
	r.HandleFunc("/spends", func(w http.ResponseWriter, r *http.Request) {
		viewsMaker.RenderView(w, "spends", nil)
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

	usersService := NewUsersService(db, graphiteClient)

	usersHandler := NewUsersHandler(usersService, loginSessionManager)
	spendingHandler := NewSpendingHandler(usersService, loginSessionManager)
	spendKindHandler := NewSpendKindHandler(db, loginSessionManager)
	debugHandler := NewDebugHandler(viewsMaker, logsPath, "ispend.log")

	// new users, list users, etc ...
	r.Handle("/users", usersHandler)
	r.Handle("/users/me/{username}/{cookie}", usersHandler)
	r.Handle("/users/login", usersHandler)
	r.Handle("/users/login/check", usersHandler)
	r.Handle("/users/logout", usersHandler)
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

	r.Use(getLoggingMiddleware(graphiteClient))
	r.Use(getPanicRecoverMiddleware(graphiteClient))

	return r
}

// TODO: make a type/struct out of this file ?
func Serve(configData []byte, port, environment string) {
	config, err := NewYamlConfig(configData)
	if err != nil {
		log.Fatalf("cannot read config file: %s", err.Error())
		return
	}

	log.Debugf("using usersService: \t\t[%s] with env [%s]", config.DBType, config.PostgresEnv)
	log.Debugf("config - usersService prod: \t%s:%d", config.DBProd.Host, config.DBProd.Port)
	log.Debugf("config - usersService dev: \t%s:%d", config.DBDev.Host, config.DBDev.Port)

	var graphiteClient *GraphiteClient
	if config.Graphite.Enabled {
		graphiteClient, err = NewGraphite(config.Graphite.Host, config.Graphite.Port)
		if err != nil {
			panicMessage := fmt.Sprintf("cannot create graphite client: %s", err.Error())
			log.Error(panicMessage)
			panic(panicMessage)
		}
		graphiteClient.Prefix = "ispend"
	} else {
		log.Debugln("using NOP graphite client")
		graphiteClient = NewGraphiteNop(config.Graphite.Host, config.Graphite.Port)
	}

	if graphiteClient.SimpleSend("stats.server.started", "1") {
		log.Traceln("stats.server.started metric successfully sent to graphite")
	} else {
		log.Errorf("failed to send stats.server.started metric to graphite")
	}

	chInterrupt := make(chan signal, 1)
	chOsInterrupt := make(chan os.Signal, 1)
	ossignal.Notify(chOsInterrupt, os.Interrupt)

	if port == "" {
		port = DefaultPort
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
	if config.DBType == DBTypePostgres {
		dbClient = NewPostgresDBClient(
			config.GetPostgresHost(),
			config.GetPostgresPort(),
			config.GetPostgresDBName(),
			config.GetPostgresDBUsername(),
			config.GetPostgresDBPassword(),
			config.GetPostgresDBSSLMode(),
			config.PingTimeout,
		)
		err = dbClient.Open()
		if err != nil {
			log.Fatalf("cannot open PS DB connection: %s", err.Error())
		}

		router = routerSetup(isProduction, dbClient, graphiteClient, chInterrupt)
		log.Debugln(" > usersService: using Postgres usersService")
	} else if config.DBType == DBTypeInMemory {
		dbClient = NewInMemoryDB()
		router = routerSetup(isProduction, dbClient, graphiteClient, chInterrupt)
		log.Debugln(" > usersService: using in memory usersService")
	} else {
		log.Fatalf("unknown usersService type from config: %s", config.DBType)
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
	gracefulShutdown(httpServer, dbClient)
}

func gracefulShutdown(httpServer *http.Server, dbClient SpenderDB) {
	log.Debug("graceful shutdown initiated ...")

	err := dbClient.Close()
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
