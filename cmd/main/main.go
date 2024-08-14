package main

import (
	"log"

	"github.com/shynn12/medods/internal/app/apiserver"
	"github.com/shynn12/medods/internal/config"
)

func main() {

	cfg := config.GetCongif()
	s := apiserver.New(cfg)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
