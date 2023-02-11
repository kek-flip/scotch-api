package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/kek-flip/scotch-api/internal/model"
	"github.com/kek-flip/scotch-api/internal/store"
)

type server struct {
	router     *mux.Router
	store      *store.Store
	err_logger *log.Logger
	logger     *log.Logger
}

func StartServer() error {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:admin@localhost:5432/scotch?sslmode=disable")
	if err != nil {
		return err
	}

	store := store.NewStore(conn)

	server := newServer(store)
	server.logger.Println("Server is listening on localhost:8080 ...")
	return http.ListenAndServe(":8080", server)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func newServer(st *store.Store) *server {
	s := &server{
		router:     mux.NewRouter(),
		store:      st,
		err_logger: newErrLogger(),
		logger:     newLogger(),
	}

	s.configRouter()

	return s
}

func newLogger() *log.Logger {
	prefix_str := "INFO:"
	return log.New(os.Stdout, prefix_str, log.LstdFlags)
}

func newErrLogger() *log.Logger {
	prefix_str := "ERROR:"
	return log.New(os.Stderr, prefix_str, log.LstdFlags)
}

func (s *server) configRouter() {
	s.router.HandleFunc("/users", s.handlerUserCreate()).Methods("POST")
}

func (s *server) handlerUserCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Started POST \"/users\"")
		s.logger.Println("Processing by heandlerUserCreate()")

		user := &model.User{}

		if err := json.NewDecoder(r.Body).Decode(user); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err)
			s.err_logger.Println("Invalid data format: ", err)
			return
		}

		if err := s.store.User().Create(user); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprint(w, err)
			s.err_logger.Println("Invalid user data: ", err)
			return
		}

		user.ClearPassword()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)

		s.logger.Printf("Completed %d CREATED\n", http.StatusCreated)
	}
}
