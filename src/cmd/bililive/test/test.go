package test

import (
	"os"
	"path/filepath"

	"github.com/hr3lxphr6j/bililive-go/src/cmd/bililive/internal/flag"
	"github.com/hr3lxphr6j/bililive-go/src/configs"
)

func getConfigBesidesExecutable() (*configs.Config, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(filepath.Dir(exePath), "config.yml")
	config, err := configs.ReadConfigWithFile(configPath)
	if err != nil {
		return nil, err
	}
	return config, nil
}
func Get_config() (*configs.Config, error) {
	var config *configs.Config
	if *flag.Conf != "" {
		c, err := configs.ReadConfigWithFile(*flag.Conf)
		if err != nil {
			return nil, err
		}
		config = c
	} else {
		config = flag.GenConfigFromFlags()
	}
	if !config.RPC.Enable && len(config.LiveRooms) == 0 {
		// if config is invalid, try using the config.yml file besides the executable file.
		config, err := getConfigBesidesExecutable()
		if err == nil {
			return config, config.Verify()
		}
	}
	return config, config.Verify()
}
