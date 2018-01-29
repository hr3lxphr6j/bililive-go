package configs

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"errors"
	"os"
	"fmt"
)

type Config struct {
	EnableRpc  bool     `yaml:"enable_rpc"`
	Interval   int      `yaml:"interval"`
	Port       uint     `yaml:"port"`
	OutPutPath string   `yaml:"out_put_path"`
	LiveRooms  []string `yaml:"live_rooms"`
}

func verifyConfig(config *Config) error {
	if config.Interval == 0 {
		return errors.New("interval can not be null or '0'")
	}
	if _, err := os.Stat(config.OutPutPath); err != nil {
		return errors.New(fmt.Sprintf(`the out put path: "%s" is not exist`, config.OutPutPath))
	}
	if config.EnableRpc {
		if config.Port == 0 {
			return errors.New("rpc listen port can not be null or '0'")
		}
	}
	return nil
}

func NewConfig(configFilePath string) (*Config, error) {
	b, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	config := new(Config)
	err = yaml.Unmarshal(b, config)
	if err != nil {
		return nil, err
	}
	if err = verifyConfig(config); err != nil {
		return nil, err
	}
	return config, nil
}
