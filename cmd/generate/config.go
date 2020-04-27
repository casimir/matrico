package main

import (
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

// common impl

type SpecConfig struct {
	Name      string
	Version   string
	Release   string
	DefsExtra []string `toml:"definitions_extra"`
	Blacklist []string
	Overrides map[string]string
}

type Config struct {
	Specs []SpecConfig
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
