package internal

import (
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
)

type LineUpdate interface {
	ID() int
}

type Line interface {
	Update(update LineUpdate)
	String() string
}

type Progress struct {
	lock           sync.Mutex
	closeOnce      sync.Once
	index          int
	out            io.Writer
	name1MaxLenght int
	name2MaxLenght int
	name3MaxLenght int
	lines          map[int]Line
	updates        chan struct{}
	done           chan struct{}
}

func NewProgress() *Progress {
	p := &Progress{
		out:     os.Stdout,
		lines:   make(map[int]Line),
		updates: make(chan struct{}),
		done:    make(chan struct{}),
	}
	go p.displayLines()
	return p
}

func (p *Progress) Close() error {
	p.closeOnce.Do(func() {
		close(p.updates)
		<-p.done
	})
	return nil
}

func (p *Progress) AddLine(line Line) int {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.index++
	p.lines[p.index] = line
	return p.index
}

func (p *Progress) updateLine(lu LineUpdate) {
	p.lines[lu.ID()].Update(lu)
	p.updates <- struct{}{}
	return
}

func (p *Progress) UpdateLine(lu LineUpdate) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.updateLine(lu)
	return
}

func (p *Progress) UpdateLines(lus ...LineUpdate) {
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, lu := range lus {
		p.updateLine(lu)
	}
	return
}

func (p *Progress) writeString(s string) (int, error) {
	return p.out.Write([]byte(s))
}

func newlines(l int) (ns string) {
	for i := 0; i < l; i++ {
		ns += "\n"
	}
	return
}

func (p *Progress) arrangedIndices() []int {
	var inds []int
	for ind := range p.lines {
		inds = append(inds, ind)
	}
	sort.Ints(inds)
	return inds
}

func (p *Progress) displayLines() {
	// Send first state update signal just to display lines.
	go func() {
		p.updates <- struct{}{}
	}()
	var nls int
	for range p.updates {
		nls = len(p.lines)
		if nls == 0 {
			continue
		}
		var s string
		for _, ind := range p.arrangedIndices() {
			s += p.lines[ind].String() + "\n"
		}
		p.writeString(s)
		p.writeString(fmt.Sprintf("\033[%dA", nls))
	}
	p.writeString(newlines(nls))
	p.done <- struct{}{}
}
