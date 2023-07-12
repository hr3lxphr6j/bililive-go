package upload

import (
	"testing"

	"github.com/matyle/bililive-go/src/configs"
)

func TestBiliUPload(t *testing.T) {
	//配置初始化
	configs.NewConfigWithFile("../../config.yaml")
	biliUps := NewBiliUPLoads(configs.NewConfig().BiliupConfigs, 2)
	biliUps.Server(RemoveFilesHandler)
}
