package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/hr3lxphr6j/bililive-go/src/api"
	"github.com/hr3lxphr6j/bililive-go/src/instance"
	"github.com/hr3lxphr6j/bililive-go/src/listeners"
	"github.com/hr3lxphr6j/bililive-go/src/recorders"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"net/url"
)

type CommonResp struct {
	ErrNo  int         `json:"err_no"`
	ErrMsg string      `json:"err_msg"`
	Data   interface{} `json:"data"`
}

type RespInfo struct {
	HostName string `json:"host_name"`
	RoomName string `json:"room_name"`
	Status   bool   `json:"status"`
}

type RespLive struct {
	Id          api.LiveId `json:"id"`
	LiveUrl     string     `json:"live_url"`
	Info        RespInfo   `json:"info"`
	IsListening bool       `json:"is_listening"`
	IsRecoding  bool       `json:"is_recoding"`
}

func live2RespLive(ctx context.Context, live api.Live) RespLive {
	inst := instance.GetInstance(ctx)
	return RespLive{
		Id:      live.GetLiveId(),
		LiveUrl: live.GetRawUrl(),
		Info: RespInfo{
			HostName: live.GetCachedInfo().HostName,
			RoomName: live.GetCachedInfo().RoomName,
			Status:   live.GetCachedInfo().Status,
		},
		IsListening: inst.ListenerManager.(listeners.IListenerManager).HasListener(ctx, api.LiveId(live.GetLiveId())),
		IsRecoding:  inst.RecorderManager.(recorders.IRecorderManager).HasRecorder(ctx, api.LiveId(live.GetLiveId())),
	}
}

func getAllLives(writer http.ResponseWriter, r *http.Request) {
	inst := instance.GetInstance(r.Context())
	lives := make([]RespLive, 0)
	for _, v := range inst.Lives {
		lives = append(lives, live2RespLive(r.Context(), v))
	}
	if resp, err := json.Marshal(CommonResp{Data: map[string]interface{}{"lives": lives}}); err == nil {
		writer.Write(resp)
	}
}

func getLives(writer http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	res := CommonResp{}
	if id, ok := vars["id"]; ok {
		inst := instance.GetInstance(r.Context())
		if live, ok2 := inst.Lives[api.LiveId(id)]; ok2 {
			res.Data = live2RespLive(r.Context(), live)
			writer.WriteHeader(http.StatusOK)
		} else {
			res.ErrNo = 404
			res.ErrMsg = fmt.Sprintf("live id: %s can not find", id)
			writer.WriteHeader(http.StatusNotFound)
		}
	} else {
		res.ErrNo = 400
		res.ErrMsg = "live id is null"
		writer.WriteHeader(http.StatusBadRequest)
	}
	if resp, err := json.Marshal(res); err == nil {
		writer.Write(resp)
	}
}

func parseLiveAction(writer http.ResponseWriter, r *http.Request) {

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
	lives := make([]RespLive, 0)
	gjson.GetBytes(b, "lives").ForEach(func(key, value gjson.Result) bool {
		isListen := value.Get("listen").Bool()
		u, _ := url.Parse(value.Get("url").String())
		live := api.NewLive(u)
		if live != nil {
			inst := instance.GetInstance(r.Context())
			if _, ok := inst.Lives[live.GetLiveId()]; !ok {
				inst.Lives[live.GetLiveId()] = live
				if isListen {
					inst.ListenerManager.(listeners.IListenerManager).AddListener(r.Context(), live)
				}
				lives = append(lives, live2RespLive(r.Context(), live))
			}
		}
		return true
	})
	resp, _ := json.Marshal(CommonResp{Data: map[string]interface{}{"lives": lives}})
	writer.Write(resp)
}
