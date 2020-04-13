package main

import (
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

// common impl

type Config struct {
	Specs []struct {
		Name      string
		Version   string
		Release   string
		Blacklist []string
	}
}

func loadConfig() (*Config, error) {
	data, err := ioutil.ReadFile("generate.toml")
	if err != nil {
		return nil, err
	}

	var config Config
	err = toml.Unmarshal(data, &config)
	return &config, err
}
