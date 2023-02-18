package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/kek-flip/scotch-api/internal/model"
	"github.com/kek-flip/scotch-api/internal/store"
)

type encd_err struct {
	Error string `json:"error"`
}

type ctxKey int

const (
	ctxUserKey ctxKey = iota
)

var (
	errWrongLoginOrPassword = errors.New("wrong login or password")
	errUnauthorized         = errors.New("not authenticated")
	errNoSuchUser           = errors.New("no user with this id")
)

type server struct {
	router       *mux.Router
	store        *store.Store
	sessionStore *sessions.CookieStore
	err_logger   *log.Logger
	logger       *log.Logger
}

func StartServer() error {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:admin@localhost:5432/scotch?sslmode=disable")
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	store := store.NewStore(conn)

	key, err := getMasterKey()
	if err != nil {
		return err
	}
	sessionStore := sessions.NewCookieStore(key)

	server := newServer(store, sessionStore)
	server.logger.Print("Server is listening on localhost:8080 ...\n\n")
	return http.ListenAndServe(":8080", server)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func getMasterKey() ([]byte, error) {
	file, err := os.OpenFile("./config/master.key", os.O_RDONLY, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	key := make([]byte, 0)

	buf := make([]byte, 8)
	for {
		n, err := file.Read(buf)
		key = append(key, buf[:n]...)
		if err == io.EOF {
			break
		}
	}

	return key, nil
}

func newServer(st *store.Store, ss *sessions.CookieStore) *server {
	s := &server{
		router:       mux.NewRouter(),
		store:        st,
		sessionStore: ss,
		err_logger:   newErrLogger(),
		logger:       newLogger(),
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
	s.router.HandleFunc("/sessions", s.heandlerSessionCreate()).Methods("POST")

	userSubrouter := s.router.PathPrefix("/users").Subrouter()
	userSubrouter.Use(s.authenticateUser)
	userSubrouter.HandleFunc("/{id:[0-9]+}", s.handlerUser()).Methods("GET")
	userSubrouter.HandleFunc("/current", s.handlerCurrentUser()).Methods("GET")
}

func (s *server) respond(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (s *server) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.sessionStore.Get(r, "scotch")
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			return
		}

		id, ok := session.Values["user_id"]
		if !ok {
			s.respond(w, http.StatusUnauthorized, encd_err{errUnauthorized.Error()})
			return
		}

		u, err := s.store.User().FindById(id.(int))
		if err != nil {
			s.respond(w, http.StatusUnauthorized, encd_err{err.Error()})
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxUserKey, u)))
	})
}

func (s *server) handlerUserCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Started POST \"/users\"")
		s.logger.Println("Processing by heandlerUserCreate()")

		user := &model.User{}

		if err := json.NewDecoder(r.Body).Decode(user); err != nil {
			s.respond(w, http.StatusBadRequest, encd_err{err.Error()})
			s.err_logger.Printf("Invalid data format: %s\n\n", err)
			return
		}

		if err := s.store.User().Create(user); err != nil {
			s.respond(w, http.StatusUnprocessableEntity, encd_err{err.Error()})
			s.err_logger.Printf("Invalid user data: %s\n\n", err)
			return
		}

		user.ClearPassword()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)

		s.logger.Printf("Completed %d CREATED\n\n", http.StatusCreated)
	}
}

func (s *server) handlerUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Started GET \"/users/:id\"")
		s.logger.Println("Processing by handlerUser()")

		vars := mux.Vars(r)

		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			s.respond(w, http.StatusBadRequest, encd_err{err.Error()})
			s.err_logger.Println("Indalid id: ", err)
			return
		}

		u, err := s.store.User().FindById(id)
		if err != nil {
			s.respond(w, http.StatusBadRequest, encd_err{errNoSuchUser.Error()})
			s.err_logger.Println("Invalid id: ", errNoSuchUser)
			return
		}

		s.respond(w, http.StatusOK, u)

		s.logger.Printf("Completed %d OK\n\n", http.StatusOK)
	}
}

func (s *server) handlerCurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Started GET \"/users/current\"")
		s.logger.Println("Processing by handlerCurrentUser()")

		u := r.Context().Value(ctxUserKey).(*model.User)

		s.respond(w, http.StatusOK, u)

		s.logger.Printf("Completed %d OK\n\n", http.StatusOK)
	}
}

func (s *server) heandlerSessionCreate() http.HandlerFunc {
	type request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Started POST \"/sessions\"")
		s.logger.Println("Processing by heandlerSessionCreate()")

		data := &request{}

		err := json.NewDecoder(r.Body).Decode(data)
		if err != nil {
			s.respond(w, http.StatusBadRequest, encd_err{err.Error()})
			s.err_logger.Printf("Invalid data format: %s\n\n", err)
			return
		}

		u, err := s.store.User().FindByLogin(data.Login)
		if err == pgx.ErrNoRows || !u.ComparePassword(data.Password) {
			s.respond(w, http.StatusUnauthorized, encd_err{errWrongLoginOrPassword.Error()})
			s.err_logger.Printf("Wrong login or password: %s\n\n", errWrongLoginOrPassword)
			return
		}

		session, err := s.sessionStore.Get(r, "scotch")
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Printf("Session error: %s\n\n", err)
			return
		}

		session.Values["user_id"] = u.ID
		if err := s.sessionStore.Save(r, w, session); err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Printf("Session error: %s\n\n", err)
			return
		}

		s.logger.Printf("Completed %d OK\n\n", http.StatusOK)
	}
}
