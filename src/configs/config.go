package configs

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// RPC info.
type RPC struct {
	Enable bool   `yaml:"enable"`
	Bind   string `yaml:"bind"`
}

var defaultRPC = RPC{
	Enable: true,
	Bind:   "127.0.0.1:8080",
}

func (r *RPC) verify() error {
	if r == nil {
		return nil
	}
	if !r.Enable {
		return nil
	}
	if _, err := net.ResolveTCPAddr("tcp", r.Bind); err != nil {
		return err
	}
	return nil
}

// Feature info.
type Feature struct {
	UseNativeFlvParser bool `yaml:"use_native_flv_parser"`
}

// VideoSplitStrategies info.
type VideoSplitStrategies struct {
	OnRoomNameChanged bool          `yaml:"on_room_name_changed"`
	MaxDuration       time.Duration `yaml:"max_duration"`
}

// Config content all config info.
type Config struct {
	RPC                  RPC      `yaml:"rpc"`
	Debug                bool     `yaml:"debug"`
	Interval             int      `yaml:"interval"`
	OutPutPath           string   `yaml:"out_put_path"`
	Feature              Feature  `yaml:"feature"`
	LiveRooms            []string `yaml:"live_rooms"`
	OutputTmpl           string   `yaml:"out_put_tmpl"`
	file                 string
	VideoSplitStrategies VideoSplitStrategies `yaml:"video_split_strategies"`
}

var defaultConfig = Config{
	RPC:        defaultRPC,
	Debug:      false,
	Interval:   30,
	OutPutPath: "./",
	Feature: Feature{
		UseNativeFlvParser: false,
	},
	LiveRooms: []string{},
	file:      "",
	VideoSplitStrategies: VideoSplitStrategies{
		OnRoomNameChanged: false,
	},
}

// Verify will return an error when this config has problem.
func (c *Config) Verify() error {
	if c == nil {
		return fmt.Errorf("config is null")
	}
	if err := c.RPC.verify(); err != nil {
		return err
	}
	if c.Interval <= 0 {
		return fmt.Errorf("the interval can not <= 0")
	}
	if _, err := os.Stat(c.OutPutPath); err != nil {
		return fmt.Errorf(`the out put path: "%s" is not exist`, c.OutPutPath)
	}
	if maxDur := c.VideoSplitStrategies.MaxDuration; maxDur > 0 && maxDur <= time.Minute {
		return fmt.Errorf("the minimum value of max_duration is one minute")
	}
	return nil
}

func NewConfigWithFile(file string) (*Config, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("can`t open file: %s", file)
	}
	config := &defaultConfig
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
