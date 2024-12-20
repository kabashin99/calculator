package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var logFile *os.File

func init() {
	var err error
	logFile, err = os.OpenFile("logs/server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.body == nil {
		lrw.body = new(bytes.Buffer)
	}
	lrw.body.Write(b)
	return lrw.ResponseWriter.Write(b)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		log.SetOutput(logFile)
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("Error reading request body:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		bodyString := string(bodyBytes)
		bodyString = strings.ReplaceAll(bodyString, "\n", "")

		log.Printf("Получен запрос: %s  %s Body: %s", r.Method, r.URL.Path, bodyString)
		next.ServeHTTP(lrw, r)
		log.Printf("Ответ: код %d Body: %s", lrw.statusCode, lrw.body.String())
	})
}

func CloseLogFile() {
	if err := logFile.Close(); err != nil {
		log.SetOutput(logFile)
		log.Printf("Error closing log file: %v", err)
	}
}
