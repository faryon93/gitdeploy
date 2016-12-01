package config

import (
	"github.com/BurntSushi/toml"
)


// ----------------------------------------------------------------------------------
//  constants
// ----------------------------------------------------------------------------------

const (
	PROVIDER_GIT = "git"
)


// ----------------------------------------------------------------------------------
//  types
// ----------------------------------------------------------------------------------

type Config struct {
	Path string
	Provider string

	Url string `toml:"url"`
	IdentityFile string `toml:"identity_file"`
	Branch string `toml:"branch"`
}


// ----------------------------------------------------------------------------------
//  functions
// ----------------------------------------------------------------------------------

func Load(path string) (*Config, error) {
	// decode the config file to struct
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}

	config.Path = path

	return &config, nil
}

