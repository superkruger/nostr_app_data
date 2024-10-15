package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name      string `yaml:"name"`
	AccountID string `yaml:"account_id"`
	Region    string `yaml:"region"`
	Branch    string `yaml:"branch"`
}

func MustNewConfig(env string) Config {
	contents, err := os.ReadFile("config/" + env + ".yaml")
	if err != nil {
		panic(err)
	}
	var c Config
	if err := yaml.Unmarshal(contents, &c); err != nil {
		panic(err)
	}
	return c
}
