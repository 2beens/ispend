package main

import (
	"flag"
	"os"
	"strings"

	"github.com/2beens/ispend"
	log "github.com/sirupsen/logrus"
)

func main() {
	displayHelp := flag.Bool("h", false, "display info/help message")
	port := flag.String("port", "", "server port")
	env := flag.String("env", "development", "set environment [development|production] [d|p]")
	logFileName := flag.String("logfile", "", "log file used to store server logs")
	logLevel := flag.String("loglvl", "", "log level")
	flag.Parse()

	if *displayHelp {
		log.Println(`
				-h                      > show this message
				-port=<port>		> used port
				-env		> environment [development|production] [d|p]
				-logfile=<logFileName>  > output log file name
				-loglvl=<logLevel>	> set log level
			`)
		log.Println()
		return
	}

	loggingSetup(*logFileName, *logLevel)

	ispend.Serve(*port, *env)
}

func loggingSetup(logFileName string, logLevel string) {
	if logLevel == "" {
		log.SetLevel(log.TraceLevel)
	} else {
		// TODO: set log level according to input string
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
