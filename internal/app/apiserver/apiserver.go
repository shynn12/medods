package apiserver

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/shynn12/medods/internal/config"
	"github.com/shynn12/medods/internal/item"
	"github.com/shynn12/medods/internal/item/db"
	"github.com/shynn12/medods/pkg/client/postgres"
	auth "github.com/shynn12/medods/pkg/jwt"
	"github.com/shynn12/medods/pkg/logging"
	"github.com/shynn12/medods/pkg/sender"
	"github.com/shynn12/medods/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type APIServer struct {
	config     *config.Config
	logger     *logging.Logger
	router     *mux.Router
	service    item.Service
	jwtManager auth.JWTManager
	sender     sender.Sender
}

func New(config *config.Config) *APIServer {
	return &APIServer{
		config: config,
		logger: logging.GetLogger(),
		router: mux.NewRouter(),
	}
}

func (s *APIServer) Start() error {

	var arg bool
	flag.BoolVar(&arg, "td", false, "Create 3 users with emails: 1@gmail.com, 2@mail.ru, 3@yandex.com")

	flag.Parse()

	s.logger.Info(s.config)

	s.logger.Info("configurating Router")
	s.configureRouter()

	s.logger.Info("configurating DB")
	err := s.configurateDB()
	if err != nil {
		log.Fatal(err)
	}

	//for testing
	if arg {
		_, err := s.service.CreateItem(context.Background(), item.ItemDTO{Email: "1@gmail.com"})
		if err != nil {
			s.logger.Error(err)
		}
		_, err = s.service.CreateItem(context.Background(), item.ItemDTO{Email: "2@mail.ru"})
		if err != nil {
			s.logger.Error(err)
		}
		_, err = s.service.CreateItem(context.Background(), item.ItemDTO{Email: "3@yandex.com"})
		if err != nil {
			s.logger.Error(err)
		}
	}

	s.logger.Info("configurate JWT")
	manager, err := auth.NewManager(s.config.Secret)
	if err != nil {
		s.logger.Fatal(err)
	}
	s.jwtManager = manager

	s.logger.Info("configurate sender")
	s.sender = sender.NewESender("testsender@mail.ru")

	s.logger.Info("starting API server")

	defer s.service.Close(context.Background())

	return http.ListenAndServe(fmt.Sprintf("%s:%s", s.config.Listen.BindIp, s.config.Listen.Port), s.router)
}

func (s *APIServer) configureRouter() {
	s.router.HandleFunc("/test", s.handleHello)
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

func (s *APIServer) handleHello(w http.ResponseWriter, r *http.Request) {

	unit, err := s.service.GetOneByGuid(context.Background(), "5")
	if err != nil {
		s.logger.Fatal(err)
	}

	res, _ := json.Marshal(unit)

	w.Write(res)

}

// Gives refresh and access tokens to users with given in URL guid
// and Update token in db (for crypting token with bcrypt (limited to 72 bytes) only part of signature will be crypted and added in base)
func (s *APIServer) handleAuth(w http.ResponseWriter, r *http.Request) {
	guid, ok := mux.Vars(r)["guid"]
	if !ok {
		s.logger.Error("Empty User GUID")
		return
	}

	access, err := s.jwtManager.NewJWT(guid, s.config.TTL.AccessTTL, r.RemoteAddr)

	if err != nil {
		s.logger.Error(err, "1")
		return
	}

	refresh, err := s.jwtManager.NewJWT(guid, s.config.TTL.RefreshTTL, r.RemoteAddr)

	if err != nil {
		s.logger.Error(err)
		return
	}
	sign := strings.Split(refresh, ".")[2][:71]

	err = s.service.UpdateToken(context.Background(), guid, sign)
	if err != nil {
		s.logger.Error(err)
		return
	}

	tokens := item.Tokens{
		Refresh: refresh,
		Access:  access,
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

// Perfom a refresh operation. Post request takes refresh token from request body
func (s *APIServer) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var tokens = &item.Tokens{}

	utils.ParseBody(r, tokens)

	refresh := tokens.Refresh

	sign := strings.Split(refresh, ".")[2][:71]

	parsedRefresh, err := s.jwtManager.Parse(refresh)
	if err != nil {
		s.logger.Error(err)
		return
	}

	//Checking expire time of refresh token
	if parsedRefresh.ExpiresAt < time.Now().Unix() {
		s.logger.Error("refresh token expired")
		w.Write([]byte("refresh token expired!"))
		return
	}

	//Get user from db
	unit, err := s.service.GetOneByGuid(context.Background(), parsedRefresh.Subject)
	if err != nil {
		s.logger.Error(err)
		return
	}

	//Validate the token
	err = bcrypt.CompareHashAndPassword([]byte(unit.Refresh), []byte(sign))
	if err != nil {
		s.logger.Errorf("wrong refresh token: %v", err)
		w.Write([]byte("wrong refresh token"))
		return
	}

	if r.RemoteAddr != parsedRefresh.IP {
		s.logger.Infof("new IP detected: %s", r.RemoteAddr)
		s.sender.SendMessage("Someone log in your account with unknown IP address", unit.Email)
	}

	//Creating new pair of tokens
	token, err := s.jwtManager.NewJWT(parsedRefresh.Subject, s.config.TTL.AccessTTL, r.RemoteAddr)

	if err != nil {
		s.logger.Error(err)
		return
	}

	refresh, err = s.jwtManager.NewJWT(parsedRefresh.Subject, s.config.TTL.RefreshTTL, r.RemoteAddr)
	if err != nil {
		s.logger.Error(err)
		return
	}

	sign = strings.Split(refresh, ".")[2][:71]

	err = s.service.UpdateToken(context.Background(), parsedRefresh.Subject, sign)
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
