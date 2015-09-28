package main

import (
	"io/ioutil"

	"github.com/pcx/example-docker-agent/log"

	"gopkg.in/yaml.v2"
)

// Config for klouds master
type ConfigScheme struct {
	UserNamespace  string `yaml:"user_namespace"`
	TokenNamespace string `yaml:"token_namespace"`
	AuthRealm      string `yaml:"auth_realm"`
	Port           string `yaml:"port"`
	MasterUser     string `yaml:"master_user"`
	MasterPass     string `yaml:"master_pass"`
	AgentURL       string `yaml:"agent_url"`
	RedisServer    string `yaml:"redis_server"`
	RedisPassword  string `yaml:"redis_password"`
	HmacKey        string `yaml:"hmac_key"`
}

var Config = ConfigScheme{}

// Init unmarshalls Config from YAML configuration in filename
func ParseConfig(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		return err
	}
	log.Infof("read config %v\n", Config)
	return nil
}
