package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/shynn12/medods/internal/config"
	"github.com/shynn12/medods/internal/item"
	"github.com/shynn12/medods/internal/item/db"
	"github.com/shynn12/medods/pkg/client/postgres"
	auth "github.com/shynn12/medods/pkg/jwt"
	"github.com/shynn12/medods/pkg/sender"
	"github.com/shynn12/medods/pkg/utils"
	"golang.org/x/crypto/bcrypt"

	"github.com/sirupsen/logrus"
)

type APIServer struct {
	config     *config.Config
	logger     *logrus.Logger
	router     *mux.Router
	service    item.Service
	jwtManager auth.JWTManager
	sender     sender.Sender
}

func New(config *config.Config) *APIServer {
	return &APIServer{
		config: config,
		logger: logrus.New(),
		router: mux.NewRouter(),
	}
}

func (s *APIServer) Start() error {
	if err := s.configureLogger(); err != nil {
		return err
	}

	s.logger.Info(s.config)

	s.logger.Info("configurating Router")
	s.configureRouter()

	s.logger.Info("configurating DB")
	s.configurateDB()

	s.logger.Info("configurate JWT")
	manager, err := auth.NewManager(s.config.Secret)
	if err != nil {
		s.logger.Fatal(err)
	}
	s.jwtManager = manager

	s.logger.Info("configurate sender")
	s.sender = sender.NewESender("testsender@mail.ru")

	s.logger.Info("starting API server")

	return http.ListenAndServe(fmt.Sprintf("%s:%s", s.config.Listen.BindIp, s.config.Listen.Port), s.router)
}

func (s *APIServer) configureLogger() error {
	level, err := logrus.ParseLevel(s.config.LogLevel)
	if err != nil {
		return err
	}

	s.logger.SetLevel(level)

	return nil
}

func (s *APIServer) configureRouter() {
	s.router.HandleFunc("/test", s.handleHello())
	s.router.HandleFunc("/refresh", s.handleRefresh).Methods(http.MethodPost)
	s.router.HandleFunc("/auth/{guid}", s.handleAuth)
}

func (s *APIServer) configurateDB() error {
	pool, err := postgres.NewClient(context.Background(), s.config.Postgres.DbURL)
	if err != nil {
		s.logger.Fatal(err)
	}

	storage := db.NewStorage(pool)

	s.service = item.NewService(storage)

	return nil
}

func (s *APIServer) handleHello() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello")
	}
}

// Give refresh and access tokens to users with given guid
func (s *APIServer) handleAuth(w http.ResponseWriter, r *http.Request) {
	guid, ok := mux.Vars(r)["guid"]
	if !ok {
		s.logger.Error("Empty User GUID")
		return
	}

	token, err := s.jwtManager.NewJWT(guid, s.config.TTL.AccessTTL, r.RemoteAddr)

	if err != nil {
		s.logger.Error(err, "1")
		return
	}

	refresh, err := s.jwtManager.NewRefreshToken()

	if err != nil {
		s.logger.Error(err)
		return
	}

	err = s.service.UpdateToken(context.Background(), guid, refresh, time.Now())
	if err != nil {
		s.logger.Error(err)
		return
	}

	tokens := item.Tokens{
		Refresh: refresh,
		Access:  token,
	}

	res, err := json.Marshal(tokens)
	if err != nil {
		s.logger.Error(err)
		return
	}

	w.Header().Set("Content-Type", "pkglication/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

// Perfom a refresh operation
func (s *APIServer) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var tokens = &item.Tokens{}
	utils.ParseBody(r, tokens)

	access, err := s.jwtManager.Parse(tokens.Access)
	if err != nil {
		s.logger.Error(err)
		return
	}
	//Get user from db
	unit, err := s.service.GetOneByGuid(context.Background(), access.Subject)
	if err != nil {
		s.logger.Error(err)
		return
	}

	//Checking expire time of refresh token
	if unit.ExpiresAt.Unix() > time.Now().Unix() {
		s.logger.Error("refresh token expired")
		w.Write([]byte("refresh token expired!"))
		return
	}

	//Validate the token
	err = bcrypt.CompareHashAndPassword([]byte(unit.Refresh), []byte(tokens.Refresh))
	if err != nil {
		s.logger.Errorf("wrong refresh token: %v", err)
		w.Write([]byte("wrong refresh token"))
		return
	}

	if r.RemoteAddr != access.IP {
		s.sender.SendMessage("Someone log in your account with unknown IP address", unit.Email)
	}

	//Creating new pair of tokens
	token, err := s.jwtManager.NewJWT(access.Subject, s.config.TTL.AccessTTL, r.RemoteAddr)

	if err != nil {
		s.logger.Error(err)
		return
	}

	refresh, err := s.jwtManager.NewRefreshToken()
	if err != nil {
		s.logger.Error(err)
		return
	}

	err = s.service.UpdateToken(context.Background(), access.Subject, refresh, time.Now())
	if err != nil {
		s.logger.Error(err)
		return
	}

	tokens = &item.Tokens{
		Refresh: refresh,
		Access:  token,
	}

	res, err := json.Marshal(tokens)
	if err != nil {
		s.logger.Error(err)
		return
	}

	w.Header().Set("Content-Type", "pkglication/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)

}
