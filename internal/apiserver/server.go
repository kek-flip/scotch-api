package apiserver

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
	sessionName        = "scotch"
	ctxUserKey  ctxKey = iota
	ctxLikeKey  ctxKey = iota
)

var (
	errWrongLoginOrPassword = errors.New("wrong login or password")
	errUnauthorized         = errors.New("you are not authenticated")
	errNoSuchUser           = errors.New("no user with this id")
	errWrongContentType     = errors.New("expected multipart")
	errEmptyUser            = errors.New("expected user data")
	errEmptyPhoto           = errors.New("expected photo")
)

type server struct {
	router       *mux.Router
	store        *store.Store
	photoStore   *store.PhotoStore
	sessionStore *sessions.CookieStore
	err_logger   *log.Logger
	logger       *log.Logger
}

func StartServer() error {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	st := store.NewStore(conn)

	photoStore, err := store.NewPhotoStore(os.Getenv("PHOTOSTORE_PATH"))
	if err != nil {
		return err
	}

	key, err := getMasterKey()
	if err != nil {
		return err
	}
	sessionStore := sessions.NewCookieStore(key)

	server := newServer(st, photoStore, sessionStore)
	server.logger.Print("Server is listening on :80 ...\n\n")
	return http.ListenAndServe(":80", server)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func getMasterKey() ([]byte, error) {
	file, err := os.OpenFile(os.Getenv("MASTERKEY_PATH"), os.O_RDONLY, 0666)
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

func newServer(st *store.Store, ps *store.PhotoStore, ss *sessions.CookieStore) *server {
	s := &server{
		router:       mux.NewRouter(),
		store:        st,
		photoStore:   ps,
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
	s.router.Use(s.logRequest)
	s.router.HandleFunc("/users", s.handlerUserCreate()).Methods("POST")
	s.router.HandleFunc("/sessions", s.handlerSessionCreate()).Methods("POST")
	s.router.HandleFunc("/users/count", s.handlerUserCount()).Methods("GET")

	userSubrouter := s.router.PathPrefix("/users").Subrouter()
	userSubrouter.Use(s.authenticateUser)
	userSubrouter.HandleFunc("/{id:[0-9]+}", s.handlerUser()).Methods("GET")
	userSubrouter.HandleFunc("/current", s.handlerCurrentUser()).Methods("GET")
	userSubrouter.HandleFunc("/current", s.handlerUserUpdate()).Methods("PATCH", "PUT")
	userSubrouter.HandleFunc("/current", s.handlerUserDelete()).Methods("DELETE")
	userSubrouter.HandleFunc("/liked", s.handlerLikedUsers()).Methods("GET")
	userSubrouter.HandleFunc("/liked_by", s.handlerLikedByUsers()).Methods("GET")
	userSubrouter.HandleFunc("/matches", s.handlerUserMathces()).Methods("GET")
	userSubrouter.HandleFunc("/count", s.handlerUserCount()).Methods("GET")
	userSubrouter.HandleFunc("/filter", s.handlerUsersByFilter()).Methods("GET")
	userSubrouter.HandleFunc("/all", s.handlerUsersAll()).Methods("GET")

	photoSubrouter := s.router.PathPrefix("/photos").Subrouter()
	photoSubrouter.Use(s.authenticateUser)
	photoSubrouter.HandleFunc("/{id:[0-9]+}", s.handlerPhoto()).Methods("GET")
	photoSubrouter.HandleFunc("/current", s.handlerCurrentUserPhoto()).Methods("GET")
	photoSubrouter.HandleFunc("/current", s.handlerPhotoUpdate()).Methods("PATCH", "PUT")

	sessionSubrouter := s.router.PathPrefix("/sessions").Subrouter()
	sessionSubrouter.Use(s.authenticateUser)
	sessionSubrouter.HandleFunc("", s.handlerSessionDelete()).Methods("DELETE")

	likeSubrouter := s.router.PathPrefix("/likes").Subrouter()
	likeSubrouter.Use(s.authenticateUser)
	likeSubrouter.Use(s.checkMatch)
	likeSubrouter.HandleFunc("", s.handlerLikeCreate()).Methods("POST")
	likeSubrouter.HandleFunc("", s.handlerLikeDelete()).Methods("DELETE")
}

func (s *server) respond(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (s *server) respondPhoto(w http.ResponseWriter, p []byte) {
	io.Copy(w, bytes.NewReader(p))
	w.Header().Add("Content-type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
}

func (s *server) authenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Authenticating user...")

		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot find session:", err.Error())
			return
		}

		id, ok := session.Values["user_id"]
		if !ok {
			s.respond(w, http.StatusUnauthorized, encd_err{errUnauthorized.Error()})
			s.err_logger.Println("Cannot find user_id:", errUnauthorized.Error())
			return
		}

		u, err := s.store.User().FindById(id.(int))
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot get user data:", err.Error())
			return
		}

		s.logger.Println("Authentication complete")
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxUserKey, u)))
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Printf("Started %s %s by %s\n", r.Method, r.RequestURI, r.RemoteAddr)

		start := time.Now()
		rw := &responceWriter{w, http.StatusOK}

		next.ServeHTTP(rw, r)

		s.logger.Printf(
			"Complited with %d %s in %v\n\n",
			rw.code,
			http.StatusText(rw.code),
			time.Since(start),
		)
	})
}

func (s *server) checkMatch(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := &model.Like{}

		if err := json.NewDecoder(r.Body).Decode(l); err != nil {
			s.respond(w, http.StatusBadRequest, encd_err{err.Error()})
			s.err_logger.Println("Invalid like data format:", err.Error())
			return
		}

		l.UserID = r.Context().Value(ctxUserKey).(*model.User).ID
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxLikeKey, l)))

		if r.Method == http.MethodDelete {
			return
		}

		if _, err := s.store.Like().FindMatchLike(l); err != nil {
			return
		}

		m := &model.Match{
			User1: l.UserID,
			User2: l.LikedUser,
		}

		if err := s.store.Match().Create(m); err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot create match:", err.Error())
			return
		}
	})
}

func (s *server) handlerUserCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by heandlerUserCreate()")

		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-type"))

		if err != nil {
			s.respond(w, http.StatusBadRequest, encd_err{err.Error()})
			s.err_logger.Println("Cannot get content-type", err.Error())
			return
		}

		if !strings.HasPrefix(mediaType, "multipart/") {
			s.respond(w, http.StatusBadRequest, encd_err{errWrongContentType.Error()})
			s.err_logger.Println("Wrong content-type:", errWrongContentType.Error())
			return
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		var up, pp bytes.Buffer
		upw := bufio.NewWriter(&up)
		ppw := bufio.NewWriter(&pp)
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
				s.err_logger.Println("Cannot parse request part:", err.Error())
				return
			}

			if strings.HasPrefix(p.Header.Get("Content-type"), "application/json") {
				io.Copy(upw, p)
			}
			if strings.HasPrefix(p.Header.Get("Content-type"), "image/jpeg") {
				io.Copy(ppw, p)
			}
		}

		if up.Len() == 0 {
			s.respond(w, http.StatusBadRequest, encd_err{errEmptyUser.Error()})
			s.err_logger.Println("Cannot create user:", errEmptyUser.Error())
			return
		}

		if up.Len() == 0 {
			s.respond(w, http.StatusBadRequest, encd_err{errEmptyPhoto.Error()})
			s.err_logger.Println("Cannot create user:", errEmptyPhoto.Error())
			return
		}

		u := &model.User{}
		if err := json.NewDecoder(&up).Decode(u); err != nil {
			s.respond(w, http.StatusBadRequest, encd_err{err.Error()})
			s.err_logger.Println("Invalid user data:", err.Error())
			return
		}
		if err := s.store.User().Create(u); err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot create user:", err.Error())
			return
		}

		if err := s.photoStore.Create(pp.Bytes(), u.ID); err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot save image:", err.Error())
			return
		}

		u.ClearPassword()

		s.respond(w, http.StatusCreated, u)
	}
}

func (s *server) handlerUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by handlerUser()")

		vars := mux.Vars(r)

		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			s.respond(w, http.StatusBadRequest, encd_err{err.Error()})
			s.err_logger.Println("Indalid id:", err.Error())
			return
		}

		u, err := s.store.User().FindById(id)
		if err != nil {
			s.respond(w, http.StatusBadRequest, encd_err{errNoSuchUser.Error()})
			s.err_logger.Println("Cannot find user:", errNoSuchUser.Error())
			return
		}

		s.respond(w, http.StatusOK, u)
	}
}

func (s *server) handlerUserUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by handlerUserUpdate()")

		userID := r.Context().Value(ctxUserKey).(*model.User).ID

		u, err := s.store.User().FindById(userID)
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{errNoSuchUser.Error()})
			s.err_logger.Println("Cannot find user:", errNoSuchUser.Error())
			return
		}

		if err := json.NewDecoder(r.Body).Decode(u); err != nil {
			s.respond(w, http.StatusBadRequest, encd_err{err.Error()})
			s.err_logger.Println("Invalid user data format:", err.Error())
			return
		}

		err = s.store.User().Update(u)
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot update user:", err.Error())
			return
		}

		s.respond(w, http.StatusOK, u)
	}
}

func (s *server) handlerCurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by handlerCurrentUser()")

		u := r.Context().Value(ctxUserKey).(*model.User)

		s.respond(w, http.StatusOK, u)
	}
}

func (s *server) handlerCurrentUserPhoto() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by handlerCurrentUserPhoto()")

		userID := r.Context().Value(ctxUserKey).(*model.User).ID
		photo, err := s.photoStore.FindById(userID)
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot find current user photo:", err.Error())
			return
		}

		s.respondPhoto(w, photo)
	}
}

func (s *server) handlerLikedUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by handlerLikesLiked()")

		userID := r.Context().Value(ctxUserKey).(*model.User).ID

		likes, err := s.store.Like().FindByUserID(userID)
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot find likes:", err.Error())
			return
		}

		users := make([]*model.User, 0)
		for _, v := range likes {
			u, err := s.store.User().FindById(v.LikedUser)
			if err != nil {
				s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
				s.err_logger.Println("Cannot find users:", err.Error())
				return
			}

			users = append(users, u)
		}

		s.respond(w, http.StatusOK, users)
	}
}

func (s *server) handlerLikedByUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by handlerLikedByUsers()")

		userID := r.Context().Value(ctxUserKey).(*model.User).ID

		likes, err := s.store.Like().FindByLikedUser(userID)
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot find likes:", err.Error())
			return
		}

		users := make([]*model.User, 0)
		for _, v := range likes {
			u, err := s.store.User().FindById(v.UserID)
			if err != nil {
				s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
				s.err_logger.Println("Cannot find users:", err.Error())
				return
			}

			users = append(users, u)
		}

		s.respond(w, http.StatusOK, users)
	}
}

func (s *server) handlerUserMathces() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by handlerUserMathces()")

		userID := r.Context().Value(ctxUserKey).(*model.User).ID

		matches, err := s.store.Match().FindByUser(userID)
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot find matches:", err.Error())
			return
		}

		users := make([]*model.User, 0)
		for _, v := range matches {
			u, err := s.store.User().FindById(v.User2)
			if err != nil {
				s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
				s.err_logger.Println("Cannot find user:", err.Error())
				return
			}

			users = append(users, u)
		}

		s.respond(w, http.StatusOK, users)
	}
}

func (s *server) handlerUserCount() http.HandlerFunc {
	type responce struct {
		UserCount int `json:"user_count"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by handlerUserCount()")

		usersCount, err := s.store.User().Count()
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot count users:", err.Error())
			return
		}

		s.respond(w, http.StatusOK, responce{UserCount: usersCount})
	}
}

func (s *server) handlerUserDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by handlerUserDelete()")

		userID := r.Context().Value(ctxUserKey).(*model.User).ID

		if err := s.store.Like().DeleteByUser(userID); err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot delete like:", err.Error())
			return
		}

		if err := s.store.Like().DeleteByLikedUser(userID); err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot delete like:", err.Error())
			return
		}

		if err := s.store.Match().DeleteByUser(userID); err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot delete match:", err.Error())
			return
		}

		if err := s.store.User().DeleteById(userID); err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot delete user:", err.Error())
			return
		}

		if err := s.photoStore.DeleteByName(strconv.Itoa(userID)); err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot delete user photo:", err.Error())
			return
		}
	}
}

func (s *server) handlerUsersByFilter() http.HandlerFunc {
	type filter struct {
		MinAge int    `json:"min_age"`
		MaxAge int    `json:"max_age"`
		Gender string `json:"gender"`
		City   string `json:"city"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by handlerUsersByFilter()")

		f := &filter{}

		queryString := r.URL.Query().Get("")

		if err := json.NewDecoder(strings.NewReader(queryString)).Decode(f); err != nil {
			s.respond(w, http.StatusBadRequest, encd_err{err.Error()})
			s.err_logger.Println("Invalid filter data format:", err.Error())
			return
		}

		userID := r.Context().Value(ctxUserKey).(*model.User).ID

		users, err := s.store.User().FindByFilters(userID, f.MinAge, f.MaxAge, f.Gender, f.City)
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot find users:", err.Error())
			return
		}

		s.respond(w, http.StatusOK, users)
	}
}

func (s *server) handlerUsersAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by handlerUsersAll()")

		userID := r.Context().Value(ctxUserKey).(*model.User).ID

		users, err := s.store.User().All(userID)
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot find users:", err.Error())
			return
		}

		s.respond(w, http.StatusOK, users)
	}
}

func (s *server) handlerPhoto() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by handlerPhoto()")

		vars := mux.Vars(r)

		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			s.respond(w, http.StatusBadRequest, encd_err{err.Error()})
			s.err_logger.Println("Indalid id:", err.Error())
			return
		}

		p, err := s.photoStore.FindById(id)
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot find user photo:", err.Error())
			return
		}

		s.respondPhoto(w, p)
	}
}

func (s *server) handlerPhotoUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by handlerPhotoUpdate()")

		p, err := io.ReadAll(r.Body)
		if err != nil {
			s.respond(w, http.StatusBadRequest, encd_err{err.Error()})
			s.err_logger.Println("Cannot read photo:", err.Error())
			return
		}

		userID := r.Context().Value(ctxUserKey).(*model.User).ID

		if err := s.photoStore.Create(p, userID); err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot update photo:", err.Error())
			return
		}
	}
}

func (s *server) handlerSessionCreate() http.HandlerFunc {
	type request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by heandlerSessionCreate()")

		data := &request{}

		if err := json.NewDecoder(r.Body).Decode(data); err != nil {
			s.respond(w, http.StatusBadRequest, encd_err{err.Error()})
			s.err_logger.Println("Invalid login and password data format:", err.Error())
			return
		}

		u, err := s.store.User().FindByLogin(data.Login)
		if err == pgx.ErrNoRows || !u.ComparePassword(data.Password) {
			s.respond(w, http.StatusUnauthorized, encd_err{errWrongLoginOrPassword.Error()})
			s.err_logger.Println("Wrong login or password:", errWrongLoginOrPassword.Error())
			return
		}

		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot get session:", err.Error())
			return
		}

		session.Values["user_id"] = u.ID
		if err := s.sessionStore.Save(r, w, session); err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot save session:", err.Error())
			return
		}
	}
}

func (s *server) handlerSessionDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by handlerSessionDelete()")

		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot get session:", err.Error())
			return
		}

		session.Options.MaxAge = -1

		if err := s.sessionStore.Save(r, w, session); err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot delete session:", err.Error())
			return
		}
	}
}

func (s *server) handlerLikeCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by heandlerLikeCreate()")

		l := r.Context().Value(ctxLikeKey).(*model.Like)

		if err := s.store.Like().Create(l); err != nil {
			s.respond(w, http.StatusInternalServerError, encd_err{err.Error()})
			s.err_logger.Println("Cannot create like:", err.Error())
			return
		}

		s.respond(w, http.StatusCreated, l)
	}
}

func (s *server) handlerLikeDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Println("Processing by handlerLikeDelete()")

		l := r.Context().Value(ctxLikeKey).(*model.Like)

		if err := s.store.Like().DeleteByUsers(l.UserID, l.LikedUser); err != nil {
			s.respond(w, http.StatusUnprocessableEntity, encd_err{err.Error()})
			s.err_logger.Println("Cannot delete like:", err.Error())
			return
		}

		if err := s.store.Match().DeleteByUser(l.UserID); err != nil {
			s.respond(w, http.StatusUnprocessableEntity, encd_err{err.Error()})
			s.err_logger.Println("Cannot delete match:", err.Error())
			return
		}
	}
}
