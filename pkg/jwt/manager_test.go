package auth_test

import (
	"testing"
	"time"

	auth "github.com/shynn12/medods/pkg/jwt"
)

func TestAuth(t *testing.T) {

	var ttl = time.Minute
	var ttl2 = time.Second
	t.Log(ttl, ttl2)
	manager, err := auth.NewManager("test")
	if err != nil {
		t.Fatal(err)
	}
	arg := auth.IpClaims{Subject: "1", ExpiresAt: time.Now().Add(ttl2).Unix(), IP: "0.0.0.0:10000"}
	access, err := manager.NewJWT(arg.Subject, ttl2, arg.IP)
	if err != nil {
		t.Fatal(err)
	}
	refresh, _ := manager.NewJWT(arg.Subject, ttl, arg.IP)

	if access == refresh {
		t.Fatal("Tokens are equal")
	}

	PAccess, err := manager.Parse(access)
	if err != nil {
		t.Fatal(err)
	}

	if *PAccess != arg {
		t.Fatal("arg != PAccess")
	}
}
