package ispend

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

type DebugHandler struct {
	logFileName string
	logFilePath string
	viewsMaker  *ViewsMaker
}

func NewDebugHandler(viewsMaker *ViewsMaker, logFilePath string, logFileName string) *DebugHandler {
	if !strings.HasSuffix(logFilePath, "/") {
		logFilePath += "/"
	}
	return &DebugHandler{
		viewsMaker:  viewsMaker,
		logFilePath: logFilePath,
		logFileName: logFileName,
	}
}

func (handler *DebugHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if r.URL.Path == "/debug" {
			handler.viewsMaker.RenderView(w, "debug", nil)
		} else if r.URL.Path == "/debug/logs" {
			handler.handleGetLogs(w, r)
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

func (handler *DebugHandler) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	logFilePath := filepath.FromSlash(handler.logFilePath + handler.logFileName)
	file, err := os.Open(logFilePath)
	if err != nil {
		_ = SendAPIErrorResp(w, "server error 10001", http.StatusInternalServerError)
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

func (handler *DebugHandler) handleUnknownPath(w http.ResponseWriter) {
	err := SendAPIErrorResp(w, "unknown path", http.StatusBadRequest)
	if err != nil {
		log.Errorf("error while sending error response to client [unknown path]: %s", err.Error())
	}
}
