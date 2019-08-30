package configs

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"gopkg.in/yaml.v2"
)

type TLS struct {
	Enable   bool   `yaml:"enable"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

func (t *TLS) Verify() error {
	if t == nil {
		return nil
	}
	if !t.Enable {
		return nil
	}
	if _, err := tls.LoadX509KeyPair(t.CertFile, t.KeyFile); err != nil {
		return err
	}
	return nil
}

type RPC struct {
	Enable bool   `yaml:"enable"`
	Bind   string `yaml:"port"`
	Token  string `yaml:"token"`
	TLS    TLS    `yaml:"tls"`
}

func (r *RPC) Verify() error {
	if r == nil {
		return nil
	}
	if !r.Enable {
		return nil
	}
	if _, err := net.ResolveTCPAddr("tcp", r.Bind); err != nil {
		return err
	}
	if err := r.TLS.Verify(); err != nil {
		return err
	}
	return nil
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

func (c *Config) Verify() error {
	if c == nil {
		return fmt.Errorf("config is null")
	}
	if err := c.RPC.Verify(); err != nil {
		return err
	}
	if c.Interval <= 0 {
		return fmt.Errorf("the interval can not <= 0")
	}
	if _, err := os.Stat(c.OutPutPath); err != nil {
		return fmt.Errorf(`the out put path: "%s" is not exist`, c.OutPutPath)
	}
	return nil
}

func NewConfigWithFile(file string) (*Config, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("can`t open file: %s", file)
	}
	config := new(Config)
	if err = yaml.Unmarshal(b, config); err != nil {
		return nil, err
	}
	config.file = file
	return config, nil
}

func (c *Config) Marshal() error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.file, b, os.ModeAppend)
}
