package listeners

type statusEvt uint8

const (
	statusToTrueEvt statusEvt = 1 << iota
	statusToFalseEvt
	roomNameChangedEvt
)

type status struct {
	roomName   string
	roomStatus bool
}

func (s status) Diff(that status) (res statusEvt) {
	if !s.roomStatus && that.roomStatus {
		res |= statusToTrueEvt
	}
	if s.roomStatus && !that.roomStatus {
		res |= statusToFalseEvt
	}
	if s.roomStatus && that.roomStatus && s.roomName != that.roomName {
		res |= roomNameChangedEvt
	}
	return res
}
