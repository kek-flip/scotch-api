package apiserver

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type server struct {
	router     *mux.Router
	err_logger *log.Logger
	logger     *log.Logger
}

func StartServer() error {
	server := newServer()
	server.logger.Println("Server is listening...")
	return http.ListenAndServe(":8080", server)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func newServer() *server {
	s := &server{
		router:     mux.NewRouter(),
		err_logger: newErrLogger(),
		logger:     newLogger(),
	}

	s.configRouter()

	return s
}

func newErrLogger() *log.Logger {
	prefix_str := "ERROR:"
	return log.New(os.Stderr, prefix_str, log.LstdFlags)
}

func newLogger() *log.Logger {
	prefix_str := "INFO:"
	return log.New(os.Stdout, prefix_str, log.LstdFlags)
}

func (s *server) configRouter() {
	// TODO: add router configuration
}
