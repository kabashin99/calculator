package middleware

import (
	"log"
	"net/http"
	"os"
)

var logFile *os.File

func init() {
	var err error
	logFile, err = os.OpenFile("logs/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.SetOutput(logFile)
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("Handled request: %s %s", r.Method, r.URL.Path)
	})
}

func CloseLogFile() {
	if err := logFile.Close(); err != nil {
		log.Printf("Error closing log file: %v", err)
	}
}
