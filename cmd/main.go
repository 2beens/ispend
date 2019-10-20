package main

import (
	"flag"
	"io/ioutil"
	"os"
	"strings"

	"github.com/2beens/ispend"
	log "github.com/sirupsen/logrus"
)

func main() {
	displayHelp := flag.Bool("h", false, "display info/help message")
	port := flag.String("port", "", "server port")
	logFileName := flag.String("logfile", "", "log file used to store server logs")
	logLevel := flag.String("loglvl", "", "log level")
	flag.Parse()

	if *displayHelp {
		log.Println(`
				-h                      > show this message
				-port=<port>		> used port
				-logfile=<logFileName>  > output log file name
				-loglvl=<logLevel>	> set log level [debug | error | fatal | info | trace | warn]
			`)
		log.Println()
		return
	}

	loggingSetup(*logFileName, *logLevel)

	yamlConfData, err := readYamlConfig()
	if err != nil {
		log.Fatalf("cannot open/read yaml conf file: %s", err.Error())
	}

	ispend.Serve(yamlConfData, *port)
}

func readYamlConfig() ([]byte, error) {
	yamlConfFile, err := os.Open("cmd/config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := yamlConfFile.Close()
		if err != nil {
			log.Errorf("read yaml config - close config error: %s", err)
		}
	}()

	return ioutil.ReadAll(yamlConfFile)
}

func loggingSetup(logFileName string, logLevel string) {
	switch strings.ToLower(logLevel) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.TraceLevel)
	}

	if logFileName == "" {
		log.SetOutput(os.Stdout)
		return
	}

	if !strings.HasSuffix(logFileName, ".log") {
		logFileName += ".log"
	}

	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		log.Panicf("failed to open log file %q: %s", logFileName, err)
	}

	log.SetOutput(logFile)
}
