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
func Get_test_Proxy() string {
	// daili := ""
	// read_config, err := getConfigBesidesExecutable()
	// if err == nil {
	// 	// fmt.Println("daili:", read_config.Proxy)
	// 	daili = read_config.Proxy
	// 	return daili
	// } else {
	// 	daili = ""
	// 	return daili
	// 	// fmt.Println("err:", err)
	// }

	// var config *configs.Config
	if *flag.Conf != "" { //从参数中加载config
		c, err := configs.ReadConfigWithFile(*flag.Conf)
		if err != nil {
			return ""
		} else {
			// config = c
			return c.Proxy
		}
	} else { //无参数，默认搜索同目录下的config.yml
		// if config is invalid, try using the config.yml file besides the executable file.
		c, err := getConfigBesidesExecutable()
		if err != nil {
			return ""
		} else {
			// config = c
			return c.Proxy
		}
	}
}
