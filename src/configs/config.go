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

// On record finished actions.
type OnRecordFinished struct {
	ConvertToMp4          bool `yaml:"convert_to_mp4"`
	DeleteFlvAfterConvert bool `yaml:"delete_flv_after_convert"`
}

// Config content all config info.
type Config struct {
	file string

	RPC                  RPC                  `yaml:"rpc"`
	Debug                bool                 `yaml:"debug"`
	Interval             int                  `yaml:"interval"`
	OutPutPath           string               `yaml:"out_put_path"`
	Feature              Feature              `yaml:"feature"`
	LiveRooms            []LiveRoom           `yaml:"live_rooms"`
	OutputTmpl           string               `yaml:"out_put_tmpl"`
	VideoSplitStrategies VideoSplitStrategies `yaml:"video_split_strategies"`
	Cookies              map[string]string    `yaml:"cookies"`
	OnRecordFinished     OnRecordFinished     `yaml:"on_record_finished"`
	TimeoutInUs          int                  `yaml:"timeout_in_us"`
}

type LiveRoom struct {
	Url         string `yaml:"url"`
	IsRecording bool   `yaml:"is_recording"`
}

type liveRoomAlias LiveRoom

// allow both string and LiveRoom format in config
func (l *LiveRoom) UnmarshalYAML(unmarshal func(interface{}) error) error {
	liveRoomAlias := liveRoomAlias{
		IsRecording: true,
	}
	if err := unmarshal(&liveRoomAlias); err != nil {
		var url string
		if err = unmarshal(&url); err != nil {
			return err
		}
		liveRoomAlias.Url = url
	}
	*l = LiveRoom(liveRoomAlias)

	return nil
}

func NewLiveRoomsWithStrings(strings []string) []LiveRoom {
	if len(strings) == 0 {
		return make([]LiveRoom, 0, 4)
	}
	liveRooms := make([]LiveRoom, len(strings))
	for index, url := range strings {
		liveRooms[index].Url = url
		liveRooms[index].IsRecording = true
	}
	return liveRooms
}

var defaultConfig = Config{
	RPC:        defaultRPC,
	Debug:      false,
	Interval:   30,
	OutPutPath: "./",
	Feature: Feature{
		UseNativeFlvParser: false,
	},
	LiveRooms: []LiveRoom{},
	file:      "",
	VideoSplitStrategies: VideoSplitStrategies{
		OnRoomNameChanged: false,
	},
	OnRecordFinished: OnRecordFinished{
		ConvertToMp4:          false,
		DeleteFlvAfterConvert: false,
	},
	TimeoutInUs: 60000000,
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
	if maxDur := c.VideoSplitStrategies.MaxDuration; maxDur > 0 && maxDur < time.Minute {
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
