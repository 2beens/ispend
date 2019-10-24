package internal

import (
	"context"
	"fmt"
	"net/http"
	"os"
	ossignal "os/signal"
	"runtime/debug"
	"strings"
	"time"

	"github.com/2beens/ispend/internal/db"
	"github.com/2beens/ispend/internal/handlers"
	"github.com/2beens/ispend/internal/metrics"
	"github.com/2beens/ispend/internal/models"
	"github.com/2beens/ispend/internal/platform"
	"github.com/2beens/ispend/internal/services"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var loginSessionManager = platform.NewLoginSessionHandler()

var dbClient db.SpenderDB

func getLoggingMiddleware(graphiteClient *metrics.GraphiteClient) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: mute path logs for now
			userAgent := r.Header.Get("User-Agent")
			sessionID := r.Header.Get("X-Ispend-SessionID")
			log.Tracef(" ====> request [%s] path: [%s] [sessionID: %s] [UA: %s]", r.Method, r.URL.Path, sessionID, userAgent)

			path := r.URL.Path
			if path == "/" {
				path = "<root>"
			} else if strings.HasPrefix(path, "/") {
				path = strings.TrimPrefix(path, "/")
			}
			graphiteClient.SimpleSendInt("paths."+path, 1)

			next.ServeHTTP(w, r)
		})
	}
}

func getPanicRecoverMiddleware(graphiteClient *metrics.GraphiteClient) func(next http.Handler) http.Handler {
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
						path = strings.TrimPrefix(path, "/")
					}
					graphiteClient.SimpleSendInt("panic.recovery."+path, 1)
				}
			}(r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

func routerSetup(logsPath string, db db.SpenderDB, graphiteClient *metrics.GraphiteClient, chInterrupt chan models.Signal) (r *mux.Router) {
	r = mux.NewRouter()

	// server static files
	fs := http.FileServer(http.Dir("./public/"))
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", fs))

	viewsMaker, err := platform.NewViewsMaker("public/views/")
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
		platform.SendAPIOKResp(w, "Oh yeah...")
	})

	r.HandleFunc("/harakiri", func(w http.ResponseWriter, r *http.Request) {
		chInterrupt <- platform.EmptySignal
		platform.SendAPIOKResp(w, "Goodbye cruel world...")
	})

	usersService := services.NewUsersService(db, graphiteClient)

	usersRouter := r.PathPrefix("/users").Subrouter()
	spendingRouter := r.PathPrefix("/spending").Subrouter()
	spendKindRouter := r.PathPrefix("/spending/kind").Subrouter()
	debugRouter := r.PathPrefix("/debug").Subrouter()
	handlers.UsersHandlerSetup(usersRouter, usersService, loginSessionManager)
	handlers.SpendingHandlerSetup(spendingRouter, usersService, loginSessionManager)
	handlers.SpendKindHandlerSetup(spendKindRouter, db, loginSessionManager)
	handlers.DebugHandlerSetup(debugRouter, viewsMaker, logsPath, "ispend.log")

	// all the rest - unknown paths
	r.HandleFunc("/{unknown}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		unknownPath := vars["unknown"]

		acceptHeader := r.Header.Get("Accept")
		log.Debugf("accept header: %s", acceptHeader)
		if strings.Contains(acceptHeader, "application/json") {
			platform.SendAPIErrorResp(w, "unknown path: "+unknownPath, http.StatusNotFound)
		} else {
			//TODO: navigate to some error page instead of silent home redirect ?
			http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		}
	})

	r.Use(getLoggingMiddleware(graphiteClient))
	r.Use(getPanicRecoverMiddleware(graphiteClient))

	return r
}

// TODO: make a type/struct out of this file ?
func Serve(configData []byte, port string) {
	dbPassword := os.Getenv("ISPEND_POSTGRESS_PASSWORD")
	if len(dbPassword) == 0 {
		log.Warn("DB password is empty string...")
	}

	config, err := platform.NewYamlConfig(configData)
	if err != nil {
		log.Fatalf("cannot read config file: %s", err.Error())
		return
	}

	log.Debugf("using usersService: \t\t[%s] with env [%s]", config.DBType, config.PostgresEnv)
	log.Debugf("config - usersService prod: \t%s:%d", config.DBProd.Host, config.DBProd.Port)
	log.Debugf("config - usersService dev: \t%s:%d", config.DBDev.Host, config.DBDev.Port)

	var graphiteClient *metrics.GraphiteClient
	if config.Graphite.Enabled {
		graphiteClient, err = metrics.NewGraphite(config.Graphite.Host, config.Graphite.Port)
		if err != nil {
			panicMessage := fmt.Sprintf("cannot create graphite client: %s", err.Error())
			log.Error(panicMessage)
			panic(panicMessage)
		}
		graphiteClient.Prefix = "ispend"
	} else {
		log.Debugln("using NOP graphite client")
		graphiteClient = metrics.NewGraphiteNop(config.Graphite.Host, config.Graphite.Port)
	}

	if graphiteClient.SimpleSend("stats.server.started", "1") {
		log.Traceln("stats.server.started metric successfully sent to graphite")
	} else {
		log.Errorf("failed to send stats.server.started metric to graphite")
	}

	chInterrupt := make(chan models.Signal, 1)
	chOsInterrupt := make(chan os.Signal, 1)
	ossignal.Notify(chOsInterrupt, os.Interrupt)

	if port == "" {
		port = platform.DefaultPort
	}

	// we need a webserver to get the pprof webserver
	pprofhost := "localhost"
	pprofport := "5002"
	go func() {
		log.Debugf("starting pprof server on [%s:%s] ...", pprofhost, pprofport)
		log.Debugln(http.ListenAndServe(pprofhost+":"+pprofport, nil))
	}()

	var router *mux.Router
	if config.DBType == platform.DBTypePostgres {
		dbClient = db.NewPostgresDBClient(
			config.GetPostgresHost(),
			config.GetPostgresPort(),
			config.GetPostgresDBName(),
			config.GetPostgresDBUsername(),
			dbPassword,
			config.GetPostgresDBSSLMode(),
			config.PingTimeout,
		)
		err = dbClient.Open()
		if err != nil {
			log.Fatalf("cannot open PS DB connection: %s", err.Error())
		}

		router = routerSetup(config.LogsPath, dbClient, graphiteClient, chInterrupt)
		log.Debugln(" > usersService: using Postgres usersService")
	} else if config.DBType == platform.DBTypeInMemory {
		dbClient = db.NewInMemoryDB()
		router = routerSetup(config.LogsPath, dbClient, graphiteClient, chInterrupt)
		log.Debugln(" > usersService: using in memory usersService")
	} else {
		log.Fatalf("unknown usersService type from config: %s", config.DBType)
	}

	ipAndPort := fmt.Sprintf("%s:%s", platform.IPAddress, port)

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

func gracefulShutdown(httpServer *http.Server, dbClient db.SpenderDB) {
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
