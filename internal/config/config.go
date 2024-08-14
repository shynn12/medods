package config

import (
	"log"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel string `yaml:"log_level"`

	Listen struct {
		Type   string `yaml:"type"`
		BindIp string `yaml:"bind_ip"`
		Port   string `yaml:"port"`
	} `yaml:"listen"`
	Postgres struct {
		DbURL string `yaml:"db_URL"`
	} `yaml:"postgres"`

	TTL struct {
		AccessTTL  time.Duration `yaml:"access"`
		RefreshTTL time.Duration `yaml:"refresh"`
	} `yaml:"ttl"`

	Secret string `yaml:"secret"`
}

var instanse *Config
var once sync.Once

func GetCongif() *Config {
	once.Do(func() {
		log.Println("read application configuration")
		instanse = &Config{}
		if err := cleanenv.ReadConfig("config.yml", instanse); err != nil {
			help, _ := cleanenv.GetDescription(instanse, nil)
			log.Println(help)
			log.Fatal(err)
		}
	})

	return instanse
}
