package internal

import (
	. "fmt"
)

type LineWithSIDAndStateUpdate struct {
	Id      int
	State   string
	Current int64
}

func (lu *LineWithSIDAndStateUpdate) ID() int {
	return lu.Id
}

type lineWithSIDAndState struct {
	sid          string
	state        string
	current      int64
	total, ratio float64
}

func NewLineWithSIDAndState(sid, state string, total int64) *lineWithSIDAndState {
	return &lineWithSIDAndState{sid: sid, state: state, total: float64(total)}
}

func (l *lineWithSIDAndState) Update(lu LineUpdate) {
	clu := lu.(*LineWithSIDAndStateUpdate)
	l.state = clu.State
	l.current = clu.Current
	l.ratio = float64(clu.Current) / l.total * 100.0
}

func (l *lineWithSIDAndState) String() string {
	return Sprintf("%9s [%s]: %6.2f%%", l.state, l.sid, l.ratio)
}
