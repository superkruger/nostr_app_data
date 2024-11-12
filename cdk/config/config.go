package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name      string `yaml:"name"`
	AccountID string `yaml:"account_id"`
	RegionA   string `yaml:"region_a"`
	RegionB   string `yaml:"region_b"`
	Branch    string `yaml:"branch"`
	DBSecret  string `yaml:"db_secret"`
	Subdomain string `yaml:"subdomain"`
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
