package instance

import (
	"sync"

	"github.com/hr3lxphr6j/bililive-go/src/api"
	"github.com/hr3lxphr6j/bililive-go/src/configs"
	"github.com/hr3lxphr6j/bililive-go/src/interfaces"
)

type Instance struct {
	WaitGroup       sync.WaitGroup          // 用于阻塞主线程，直到所有模块均结束
	Config          *configs.Config         // 配置信息
	Logger          *interfaces.Logger      // Log
	Lives           map[api.LiveId]api.Live // 直播间集合
	Server          interfaces.Module       // RPC服务
	EventDispatcher interfaces.Module       // 事件分发
	ListenerManager interfaces.Module       // 直播间状态监听
	RecorderManager interfaces.Module       // 录制模块
}
