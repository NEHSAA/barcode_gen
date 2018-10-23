package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

var Config struct {
	Secrets struct {
		Oauth struct {
			ClientID     string `yaml:"client_id"`
			ClientSecret string `yaml:"client_secret"`
		} `yaml:"oauth"`
	}
}

func InitConfig() error {
	data, err := ioutil.ReadFile("secrets.yaml")
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, &Config.Secrets)
	if err != nil {
		return err
	}

	fmt.Printf("loaded config: %v\n", Config)
	return nil
}
