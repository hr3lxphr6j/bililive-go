package live

import (
	"encoding/json"
)

type Info struct {
	Live                        Live
	HostName, RoomName          string
	Status, Listening, Recoding bool
}

func (i *Info) MarshalJSON() ([]byte, error) {
	t := struct {
		Id                ID     `json:"id"`
		LiveUrl           string `json:"live_url"`
		PlatformCNName    string `json:"platform_cn_name"`
		HostName          string `json:"host_name"`
		RoomName          string `json:"room_name"`
		Status            bool   `json:"status"`
		Listening         bool   `json:"listening"`
		Recoding          bool   `json:"recoding"`
		LastStartTime     string `json:"last_start_time,omitempty"`
		LastStartTimeUnix int64  `json:"last_start_time_unix,omitempty"`
	}{
		Id:             i.Live.GetLiveId(),
		LiveUrl:        i.Live.GetRawUrl(),
		PlatformCNName: i.Live.GetPlatformCNName(),
		HostName:       i.HostName,
		RoomName:       i.RoomName,
		Status:         i.Status,
		Listening:      i.Listening,
		Recoding:       i.Recoding,
	}
	if !i.Live.GetLastStartTime().IsZero() {
		t.LastStartTime = i.Live.GetLastStartTime().Format("2006-01-02 15:04:05")
		t.LastStartTimeUnix = i.Live.GetLastStartTime().Unix()
	}
	return json.Marshal(t)
}
