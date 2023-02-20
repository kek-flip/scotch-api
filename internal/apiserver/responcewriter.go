package apiserver

import "net/http"

type responceWriter struct {
	http.ResponseWriter
	code int
}

func (rw *responceWriter) WriteHeader(statusCode int) {
	rw.code = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
