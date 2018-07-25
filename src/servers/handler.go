package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/tidwall/gjson"

	"github.com/hr3lxphr6j/bililive-go/src/api"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/listeners"
	"github.com/hr3lxphr6j/bililive-go/src/recorders"
)

type CommonResp struct {
	ErrNo  int         `json:"err_no"`
	ErrMsg string      `json:"err_msg"`
	Data   interface{} `json:"data"`
}

func parseInfo(ctx context.Context, live api.Live) *api.Info {
	inst := instance.GetInstance(ctx)
	info := live.GetCachedInfo()
	info.Listening = inst.ListenerManager.(listeners.IListenerManager).HasListener(ctx, api.LiveId(live.GetLiveId()))
	info.Recoding = inst.RecorderManager.(recorders.IRecorderManager).HasRecorder(ctx, api.LiveId(live.GetLiveId()))
	return info
}

func getAllLives(writer http.ResponseWriter, r *http.Request) {
	inst := instance.GetInstance(r.Context())
	info := make([]*api.Info, 0)
	for _, v := range inst.Lives {
		info = append(info, parseInfo(r.Context(), v))
	}
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
	if live, ok := inst.Lives[api.LiveId(vars["id"])]; ok {
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
	if live, ok := inst.Lives[api.LiveId(vars["id"])]; ok {
		switch vars["action"] {
		case "start":
			inst.ListenerManager.(listeners.IListenerManager).AddListener(r.Context(), live)
			resp.Data = parseInfo(r.Context(), live)
		case "stop":
			inst.ListenerManager.(listeners.IListenerManager).RemoveListener(r.Context(), live.GetLiveId())
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
	info := make([]*api.Info, 0)
	gjson.GetBytes(b, "lives").ForEach(func(key, value gjson.Result) bool {
		isListen := value.Get("listen").Bool()
		u, _ := url.Parse(value.Get("url").String())
		if live, err := api.NewLive(u); err == nil {
			inst := instance.GetInstance(r.Context())
			if _, ok := inst.Lives[live.GetLiveId()]; !ok {
				inst.Lives[live.GetLiveId()] = live
				if isListen {
					inst.ListenerManager.(listeners.IListenerManager).AddListener(r.Context(), live)
				}
				info = append(info, parseInfo(r.Context(), live))
			}
		}
		return true
	})
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
	if live, ok := inst.Lives[api.LiveId(vars["id"])]; ok {
		inst.ListenerManager.(listeners.IListenerManager).RemoveListener(r.Context(), live.GetLiveId())
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
