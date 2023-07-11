package configs

import "github.com/matyle/bililive-go/src/live"

type liveRoomAlias LiveRoom

type LiveRoom struct {
	Url         string  `yaml:"url"`
	IsListening bool    `yaml:"is_listening"`
	LiveId      live.ID `yaml:"-"`
	Quality     int     `yaml:"quality"`
}

func NewLiveRoomsWithStrings(strings []string) []LiveRoom {
	if len(strings) == 0 {
		return make([]LiveRoom, 0, 4)
	}
	liveRooms := make([]LiveRoom, len(strings))
	for index, url := range strings {
		liveRooms[index].Url = url
		liveRooms[index].IsListening = true
		liveRooms[index].Quality = 0
	}
	return liveRooms
}
