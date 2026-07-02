package middleware

import (
	"log"
	"net/http"
	"time"
)

type ResponseWriterWrapper  struct {
	original http.ResponseWriter
	status   int
}

func (w *ResponseWriterWrapper ) Header() http.Header {
	return w.original.Header()
}

func (w *ResponseWriterWrapper ) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	return w.original.Write(b)
}

func (w *ResponseWriterWrapper ) WriteHeader(statusCode int) {
	w.status = statusCode
	w.original.WriteHeader(w.status)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		beginTime := time.Now()
		wrapper := &ResponseWriterWrapper {
			original: w,
			status:   0,
		}
		next.ServeHTTP(wrapper, r)
		runTime := time.Since(beginTime)

		log.Printf("[%v] %v %v (%v)", r.Method, r.URL.Path, wrapper.status, runTime.Abs().String())
	})
}
