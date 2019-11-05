package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"

	"github.com/gorilla/mux"
	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/consts"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/listeners"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/recorders"
)

type CommonResp struct {
	ErrNo  int         `json:"err_no"`
	ErrMsg string      `json:"err_msg"`
	Data   interface{} `json:"data"`
}

func parseInfo(ctx context.Context, l live.Live) *live.Info {
	inst := instance.GetInstance(ctx)
	obj, _ := inst.Cache.Get(l)
	info := obj.(*live.Info)
	info.Listening = inst.ListenerManager.(listeners.Manager).HasListener(ctx, l.GetLiveId())
	info.Recoding = inst.RecorderManager.(recorders.Manager).HasRecorder(ctx, l.GetLiveId())
	return info
}

type liveSlice []*live.Info

func (c liveSlice) Len() int {
	return len(c)
}
func (c liveSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c liveSlice) Less(i, j int) bool {
	return c[i].Live.GetLiveId() < c[j].Live.GetLiveId()
}

func getAllLives(writer http.ResponseWriter, r *http.Request) {
	inst := instance.GetInstance(r.Context())
	info := liveSlice(make([]*live.Info, 0))
	for _, v := range inst.Lives {
		info = append(info, parseInfo(r.Context(), v))
	}
	sort.Sort(info)
	resp := CommonResp{
		Data: info,
	}
	if b, err := json.Marshal(resp); err == nil {
		writer.Write(b)
	}
}

func getLive(writer http.ResponseWriter, r *http.Request) {
	inst := instance.GetInstance(r.Context())
	vars := mux.Vars(r)
	resp := CommonResp{}
	if live, ok := inst.Lives[live.ID(vars["id"])]; ok {
		resp.Data = parseInfo(r.Context(), live)
	} else {
		resp.ErrNo = 404
		resp.ErrMsg = fmt.Sprintf("live id: %s can not find", vars["id"])
		writer.WriteHeader(http.StatusNotFound)
	}
	if resp, err := json.Marshal(resp); err == nil {
		writer.Write(resp)
	}
}

func parseLiveAction(writer http.ResponseWriter, r *http.Request) {
	inst := instance.GetInstance(r.Context())
	vars := mux.Vars(r)
	resp := CommonResp{}
	if live, ok := inst.Lives[live.ID(vars["id"])]; ok {
		switch vars["action"] {
		case "start":
			inst.ListenerManager.(listeners.Manager).AddListener(r.Context(), live)
			resp.Data = parseInfo(r.Context(), live)
		case "stop":
			inst.ListenerManager.(listeners.Manager).RemoveListener(r.Context(), live.GetLiveId())
			resp.Data = parseInfo(r.Context(), live)
		default:
			resp.ErrNo = 400
			resp.ErrMsg = fmt.Sprintf("invalid Action: %s", vars["action"])
			writer.WriteHeader(http.StatusBadRequest)
		}
	} else {
		resp.ErrNo = 404
		resp.ErrMsg = fmt.Sprintf("live id: %s can not find", vars["id"])
		writer.WriteHeader(http.StatusNotFound)
	}
	if resp, err := json.Marshal(resp); err == nil {
		writer.Write(resp)
	}

}

/* Post data example
{
    "lives": [
        {
            "url": "http://live.bilibili.com/1030",
            "listen": true
        },
        {
            "url": "https://live.bilibili.com/493",
            "listen": true
        }
    ]
}
*/
func addLives(writer http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	info := liveSlice(make([]*live.Info, 0))
	gjson.GetBytes(b, "lives").ForEach(func(key, value gjson.Result) bool {
		isListen := value.Get("listen").Bool()
		u, _ := url.Parse(value.Get("url").String())
		if live, err := live.New(u, instance.GetInstance(r.Context()).Cache); err == nil {
			inst := instance.GetInstance(r.Context())
			if _, ok := inst.Lives[live.GetLiveId()]; !ok {
				inst.Lives[live.GetLiveId()] = live
				if isListen {
					inst.ListenerManager.(listeners.Manager).AddListener(r.Context(), live)
				}
				info = append(info, parseInfo(r.Context(), live))
			}
		}
		return true
	})
	sort.Sort(info)
	resp := CommonResp{
		Data: info,
	}
	if resp, err := json.Marshal(resp); err == nil {
		writer.Write(resp)
	}
}

func getConfig(writer http.ResponseWriter, r *http.Request) {
	resp := CommonResp{
		Data: instance.GetInstance(r.Context()).Config,
	}
	if resp, err := json.Marshal(resp); err == nil {
		writer.Write(resp)
	}
}

func putConfig(writer http.ResponseWriter, r *http.Request) {
	resp := CommonResp{}
	configRoom := instance.GetInstance(r.Context()).Config.LiveRooms
	configRoom = make([]string, 0)
	for _, live := range instance.GetInstance(r.Context()).Lives {
		configRoom = append(configRoom, live.GetRawUrl())
	}
	instance.GetInstance(r.Context()).Config.LiveRooms = configRoom
	if err := instance.GetInstance(r.Context()).Config.Marshal(); err == nil {
		resp.Data = "OK"
	} else {
		resp.ErrNo = 400
		resp.ErrMsg = err.Error()
		writer.WriteHeader(http.StatusBadRequest)
	}
	if resp, err := json.Marshal(resp); err == nil {
		writer.Write(resp)
	}
}

func removeLive(writer http.ResponseWriter, r *http.Request) {
	inst := instance.GetInstance(r.Context())
	vars := mux.Vars(r)
	res := CommonResp{}
	if live, ok := inst.Lives[live.ID(vars["id"])]; ok {
		inst.ListenerManager.(listeners.Manager).RemoveListener(r.Context(), live.GetLiveId())
		delete(inst.Lives, live.GetLiveId())
		res.Data = "OK"
	} else {
		res.ErrNo = 404
		res.ErrMsg = fmt.Sprintf("live id: %s can not find", vars["id"])
		writer.WriteHeader(http.StatusNotFound)
	}
	if resp, err := json.Marshal(res); err == nil {
		writer.Write(resp)
	}
}

func getInfo(writer http.ResponseWriter, r *http.Request) {
	res := CommonResp{
		Data: consts.AppInfo,
	}
	if resp, err := json.Marshal(res); err == nil {
		writer.Write(resp)
	}
}
