package auth

import (
	"fmt"
	"time"
)

type IpClaims struct {
	ExpiresAt int64  `json:"exp,omitempty"`
	Subject   string `json:"sub,omitempty"`
	IP        string `json:"ip,omitempty"`
}

func (c IpClaims) Valid() error {
	now := time.Now().Unix()

	if now < c.ExpiresAt {
		delta := now - c.ExpiresAt
		return fmt.Errorf("token is expired by %v", delta)
	}

	return nil
}
