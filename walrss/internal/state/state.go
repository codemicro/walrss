package state

import (
	"errors"
	"fmt"
	"github.com/kkyr/fig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	bh "github.com/timshannon/bolthold"
	"io/ioutil"
	"os"
	"strings"
)

type State struct {
	Config *Config
	Data   *bh.Store
}

func New() *State {
	return &State{}
}

type Config struct {
	Email struct {
		Host     string `fig:"host" validate:"required"`
		Username string `fig:"username" validate:"required"`
		Password string `fig:"password" validate:"required"`
		From     string `fig:"from" validate:"required"`
		Port     int    `fig:"port" validate:"required"`
	}
	Server struct {
		Host        string `fig:"host" default:"127.0.0.1"`
		Port        int    `fig:"port" default:"8080"`
		ExternalURL string `fig:"externalURL" validate:"required"`
	}
	Platform struct {
		DisableRegistration  bool `fig:"disableRegistration"`
		DisableSecureCookies bool `fig:"disableSecureCookies"`
	}
	Debug bool `fig:"debug"`
}

const configFilename = "config.yaml"

func LoadConfig() (*Config, error) {
	// If the file doesn't exist, Fig will throw a hissy fit, so we should create a blank one if it doesn't exist
	if _, err := os.Stat(configFilename); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// If the file doesn't have contents, Fig will throw an EOF, despite `touch config.yaml` working fine. idk lol
			if err := ioutil.WriteFile(configFilename, []byte("{}"), 0777); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	cfg := new(Config)
	if err := fig.Load(cfg); err != nil {
		return nil, err
	}

	cfg.Server.ExternalURL = strings.TrimSuffix(cfg.Server.ExternalURL, "/")

	if !cfg.Debug {
		log.Logger = log.Logger.Level(zerolog.InfoLevel)
	}

	return cfg, nil
}

func (cfg *Config) GetHTTPAddress() string {
	return fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
}

func (cfg *Config) EnableSecureCookies() bool {
	if cfg.Debug {
		return false
	}
	return !cfg.Platform.DisableSecureCookies
}
