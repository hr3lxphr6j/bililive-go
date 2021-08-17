package servers

import (
	"context"
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

// FIXME: remove this
func parseInfo(ctx context.Context, l live.Live) *live.Info {
	inst := instance.GetInstance(ctx)
	obj, _ := inst.Cache.Get(l)
	info := obj.(*live.Info)
	info.Listening = inst.ListenerManager.(listeners.Manager).HasListener(ctx, l.GetLiveId())
	info.Recoding = inst.RecorderManager.(recorders.Manager).HasRecorder(ctx, l.GetLiveId())
	return info
}

func getAllLives(writer http.ResponseWriter, r *http.Request) {
	inst := instance.GetInstance(r.Context())
	lives := liveSlice(make([]*live.Info, 0, 4))
	for _, v := range inst.Lives {
		lives = append(lives, parseInfo(r.Context(), v))
	}
	sort.Sort(lives)
	writeJSON(writer, lives)
}

func getLive(writer http.ResponseWriter, r *http.Request) {
	inst := instance.GetInstance(r.Context())
	vars := mux.Vars(r)
	live, ok := inst.Lives[live.ID(vars["id"])]
	if !ok {
		writeMsg(writer, http.StatusNotFound, fmt.Sprintf("live id: %s can not find", vars["id"]))
		return
	}
	writeJSON(writer, parseInfo(r.Context(), live))
}

func parseLiveAction(writer http.ResponseWriter, r *http.Request) {
	inst := instance.GetInstance(r.Context())
	vars := mux.Vars(r)
	live, ok := inst.Lives[live.ID(vars["id"])]
	if !ok {
		writeMsg(writer, http.StatusNotFound, fmt.Sprintf("live id: %s can not find", vars["id"]))
		return
	}
	switch vars["action"] {
	case "start":
		if err := inst.ListenerManager.(listeners.Manager).AddListener(r.Context(), live); err != nil {
			writeMsg(writer, http.StatusBadRequest, err.Error())
			return
		}
	case "stop":
		if err := inst.ListenerManager.(listeners.Manager).RemoveListener(r.Context(), live.GetLiveId()); err != nil {
			writeMsg(writer, http.StatusBadRequest, err.Error())
			return
		}
	default:
		writeMsg(writer, http.StatusBadRequest, fmt.Sprintf("invalid Action: %s", vars["action"]))
		return
	}
	writeJSON(writer, parseInfo(r.Context(), live))
}

/* Post data example
[
	{
		"url": "http://live.bilibili.com/1030",
		"listen": true
	},
	{
		"url": "https://live.bilibili.com/493",
		"listen": true
	}
]
*/
func addLives(writer http.ResponseWriter, r *http.Request) {
	b, _ := ioutil.ReadAll(r.Body)
	info := liveSlice(make([]*live.Info, 0))
	gjson.ParseBytes(b).ForEach(func(key, value gjson.Result) bool {
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
	writeJSON(writer, info)
}

func getConfig(writer http.ResponseWriter, r *http.Request) {
	writeJSON(writer, instance.GetInstance(r.Context()).Config)
}

func putConfig(writer http.ResponseWriter, r *http.Request) {
	configRoom := instance.GetInstance(r.Context()).Config.LiveRooms
	configRoom = make([]string, 0, 4)
	for _, live := range instance.GetInstance(r.Context()).Lives {
		configRoom = append(configRoom, live.GetRawUrl())
	}
	instance.GetInstance(r.Context()).Config.LiveRooms = configRoom
	if err := instance.GetInstance(r.Context()).Config.Marshal(); err != nil {
		writeMsg(writer, http.StatusBadRequest, err.Error())
		return
	}
	writeMsg(writer, http.StatusOK, "OK")
}

func removeLive(writer http.ResponseWriter, r *http.Request) {
	inst := instance.GetInstance(r.Context())
	vars := mux.Vars(r)
	live, ok := inst.Lives[live.ID(vars["id"])]
	if !ok {
		writeMsg(writer, http.StatusNotFound, fmt.Sprintf("live id: %s can not find", vars["id"]))
		return
	}
	lm := inst.ListenerManager.(listeners.Manager)
	if lm.HasListener(r.Context(), live.GetLiveId()) {
		if err := lm.RemoveListener(r.Context(), live.GetLiveId()); err != nil {
			writeMsg(writer, http.StatusBadRequest, err.Error())
			return
		}
	}
	delete(inst.Lives, live.GetLiveId())
	writeMsg(writer, http.StatusOK, "OK")
}

func getInfo(writer http.ResponseWriter, r *http.Request) {
	writeJSON(writer, consts.AppInfo)
}
