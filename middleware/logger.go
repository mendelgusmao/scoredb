package middleware

import (
	"fmt"
	"log"
	"net/http"
)

const (
	logFormat            = "%s - \"%s %s\" %d"
	logChannelBufferSize = 256
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

type Logger struct {
	next    http.Handler
	channel chan string
}

func NewLogger(next http.Handler) *Logger {
	logger := &Logger{
		next:    next,
		channel: make(chan string, logChannelBufferSize),
	}

	go logger.printer()

	return logger
}

func (m *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	recorder := &statusRecorder{w, http.StatusOK}

	m.next.ServeHTTP(recorder, r)

	m.emit(r.RemoteAddr, r.Method, r.RequestURI, recorder.status)
}

func (m *Logger) emit(remoteAddr, method, requestURI string, status int) {
	m.channel <- fmt.Sprintf(logFormat, remoteAddr, method, requestURI, status)
}

func (m *Logger) printer() {
	for {
		select {
		case message := <-m.channel:
			log.Println(message)
		}
	}
}
