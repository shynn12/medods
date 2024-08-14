package auth

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWTManager interface {
	NewJWT(id string, ttl time.Duration, ip string) (string, error)
	Parse(accessToken string) (*IpClaims, error)
	NewRefreshToken() (string, error)
}

type Manager struct {
	secret string
}

func NewManager(secret string) (*Manager, error) {
	if secret == "" {
		return nil, fmt.Errorf("empty signing key")
	}

	return &Manager{secret: secret}, nil
}

func (m *Manager) NewJWT(id string, ttl time.Duration, ip string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, IpClaims{
		Subject:   id,
		IP:        ip,
		ExpiresAt: time.Now().Add(ttl).Unix(),
	})

	tokenString, err := token.SignedString([]byte(m.secret))
	if err != nil {
		return "", err
	}

	return tokenString, err
}

func (m *Manager) Parse(accessToken string) (*IpClaims, error) {
	token, err := jwt.Parse(accessToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Header["alg"])
		}

		return []byte(m.secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("error get user claims from token")
	}

	res := &IpClaims{
		ExpiresAt: int64(claims["exp"].(float64)),
		Subject:   claims["sub"].(string),
		IP:        claims["ip"].(string),
	}

	return res, nil
}

func (M *Manager) NewRefreshToken() (string, error) {
	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	_, err := r.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
