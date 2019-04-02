package configs

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type TLS struct {
	Enable   bool   `yaml:"enable"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}
type RPC struct {
	Enable bool   `yaml:"enable"`
	Port   string `yaml:"port"`
	Token  string `yaml:"token"`
	TLS    TLS    `yaml:"tls"`
}
type Feature struct {
	UseNativeFlvParser bool `yaml:"use_native_flv_parser"`
}
type Config struct {
	RPC        RPC      `yaml:"rpc"`
	Debug      bool     `yaml:"debug"`
	Interval   int      `yaml:"interval"`
	OutPutPath string   `yaml:"out_put_path"`
	Feature    Feature  `yaml:"feature"`
	LiveRooms  []string `yaml:"live_rooms"`
	file       string
}

func VerifyConfig(config *Config) error {
	if config.Interval <= 0 {
		return errors.New(fmt.Sprintf(`the interval can not <= 0`))
	}
	if _, err := os.Stat(config.OutPutPath); err != nil {
		return errors.New(fmt.Sprintf(`the out put path: "%s" is not exist`, config.OutPutPath))
	}
	if config.RPC.Enable {
		if config.RPC.Port == "" {
			return errors.New("rpc listen port can not be null")
		}
		if config.RPC.TLS.Enable {
			if _, err := tls.LoadX509KeyPair(config.RPC.TLS.CertFile, config.RPC.TLS.KeyFile); err != nil {
				return err
			}
		}
	}
	return nil
}

func NewConfigWithFile(configFilePath string) (*Config, error) {
	b, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("can`t open file: %s", configFilePath))
	}
	config := new(Config)
	err = yaml.Unmarshal(b, config)
	if err != nil {
		return nil, err
	}
	config.file = configFilePath
	return config, nil
}

func (config *Config) Marshal() error {
	b, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(config.file, b, os.ModeAppend)
}
