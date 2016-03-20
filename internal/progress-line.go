package internal

import (
	. "fmt"
)

type ProgressLineUpdate struct {
	Id      int
	State   string
	Current int64
	Total   int64
}

func (lu *ProgressLineUpdate) ID() int {
	return lu.Id
}

// Line for displaying generic progress.
type progressLine struct {
	sid                   string
	state                 string
	current, total, ratio float64
}

func NewProgressLine(sid string) *progressLine {
	return &progressLine{sid: sid}
}

func (l *progressLine) Update(lu LineUpdate) {
	plu := lu.(*ProgressLineUpdate)
	l.state = plu.State
	l.current = float64(plu.Current)
	l.total = float64(plu.Total)
	l.ratio = l.current / l.total * 100.0
}

func (l *progressLine) String() string {
	return Sprintf("%9s [%s]: %6.2f%%", l.state, l.sid, l.ratio)
}
