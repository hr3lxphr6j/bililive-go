package internal

import (
	"fmt"
	"net/url"
	"time"

	"github.com/hr3lxphr6j/bililive-go/src/lib/utils"
	"github.com/hr3lxphr6j/bililive-go/src/live"
)

var livePlatformCNNameMap = map[string]string{
	"live.bilibili.com": "哔哩哔哩",
	"www.zhanqi.tv":     "战旗",
	"www.douyu.com":     "斗鱼",
	"star.longzhu.com":  "龙珠",
	"www.huomao.com":    "火猫",
	"www.yizhibo.com":   "一直播",
	"www.twitch.tv":     "twitch",
	"www.huya.com":      "虎牙",
	"cc.163.com":        "CC直播",
	"www.openrec.tv":    "openrec",
	"egame.qq.com":      "企鹅电竞",
}

type AbstractLive struct {
	Url           *url.URL
	LastStartTime time.Time
	LiveId        live.ID
}

func genLiveId(url *url.URL) live.ID {
	return live.ID(utils.GetMd5String([]byte(fmt.Sprintf("%s%s", url.Host, url.Path))))
}

func NewAbstractLive(url *url.URL) AbstractLive {
	return AbstractLive{
		Url:    url,
		LiveId: genLiveId(url),
	}
}

func (a *AbstractLive) GetLiveId() live.ID {
	return a.LiveId
}

func (a *AbstractLive) GetRawUrl() string {
	return a.Url.String()
}

func (a *AbstractLive) GetPlatformCNName() string {
	// TODO: fix this
	return livePlatformCNNameMap[a.Url.Host]
}

func (a *AbstractLive) GetLastStartTime() time.Time {
	return a.LastStartTime
}

func (a *AbstractLive) SetLastStartTime(time time.Time) {
	a.LastStartTime = time
}
