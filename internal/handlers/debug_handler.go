package handlers

import (
	"io"
	"net/http"
	"os"

	"github.com/2beens/ispend/internal/platform"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type DebugHandler struct {
	viewsMaker *platform.ViewsMaker
	logFile    string
}

func DebugHandlerSetup(router *mux.Router, viewsMaker *platform.ViewsMaker, logFile string) {
	handler := &DebugHandler{
		viewsMaker: viewsMaker,
		logFile:    logFile,
	}

	router.HandleFunc("", handler.handleGetDebugPage)
	router.HandleFunc("/logs", handler.handleGetLogs)
}

func (handler *DebugHandler) handleGetDebugPage(w http.ResponseWriter, r *http.Request) {
	handler.viewsMaker.RenderView(w, "debug", nil)
}

func (handler *DebugHandler) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(handler.logFile)
	if err != nil {
		platform.SendAPIErrorResp(w, "server error 10001", http.StatusInternalServerError)
		log.Errorf("error [%s]: %s", r.URL.Path, err.Error())
		return
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Errorf("error closing log file %s: %s", r.URL.Path, err.Error())
		}
	}()

	var logContent []byte
	readBuffer := make([]byte, 32*1024)
	for {
		n, err := file.Read(readBuffer)
		if n > 0 {
			nextB := readBuffer[:n]
			logContent = append(logContent, nextB...)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("read %d bytes: %v", n, err)
			break
		}
	}

	_, err = w.Write(logContent)
	if err != nil {
		log.Errorf("error sending log file %s: %s", r.URL.Path, err.Error())
	}
}
