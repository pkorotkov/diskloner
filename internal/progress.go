package internal

import (
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
)

type bar struct {
	Name1, Name2, Name3 string
	Current, Total      int64
	Ratio               float64
}

type BarUpdate struct {
	ID      int
	Current int64
}

type Progress struct {
	lock           sync.Mutex
	closeOnce      sync.Once
	index          int
	out            io.Writer
	name1MaxLenght int
	name2MaxLenght int
	name3MaxLenght int
	bars           map[int]*bar
	updates        chan struct{}
	done           chan struct{}
}

func NewProgress() *Progress {
	p := &Progress{
		out:     os.Stdout,
		bars:    make(map[int]*bar),
		updates: make(chan struct{}),
		done:    make(chan struct{}),
	}
	go p.display()
	return p
}

func (p *Progress) Close() error {
	p.closeOnce.Do(func() {
		close(p.updates)
		<-p.done
	})
	return nil
}

func (p *Progress) AddBar(name1, name2, name3 string, total int64) int {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.index++
	p.bars[p.index] = &bar{name1, name2, name3, 0, total, 0.0}
	var l int
	if l = len(name1); p.name1MaxLenght < l {
		p.name1MaxLenght = l
	}
	if l = len(name2); p.name2MaxLenght < l {
		p.name2MaxLenght = l
	}
	if l = len(name3); p.name3MaxLenght < l {
		p.name3MaxLenght = l
	}
	return p.index
}

func (p *Progress) HasBar(bid int) bool {
	p.lock.Lock()
	defer p.lock.Unlock()
	_, ok := p.bars[bid]
	return ok
}

func (p *Progress) updateBar(bu BarUpdate) {
	p.bars[bu.ID].Current = bu.Current
	p.bars[bu.ID].Ratio = float64(bu.Current) / float64(p.bars[bu.ID].Total) * 100
	p.updates <- struct{}{}
	return
}

func (p *Progress) UpdateBar(bu BarUpdate) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.updateBar(bu)
	return
}

func (p *Progress) UpdateBars(bus ...BarUpdate) {
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, bu := range bus {
		p.updateBar(bu)
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
	for ind := range p.bars {
		inds = append(inds, ind)
	}
	sort.Ints(inds)
	return inds
}

func (p *Progress) display() {
	// Send first state update signal just to display bars.
	go func() {
		p.updates <- struct{}{}
	}()
	var nbars int
	for range p.updates {
		nbars = len(p.bars)
		if nbars == 0 {
			continue
		}
		var s string
		fs := fmt.Sprintf("[%%%ds][%%%ds]{%%%ds}: %%.2f%%%% (%%d / %%d)\n", p.name1MaxLenght, p.name2MaxLenght, p.name3MaxLenght)
		for _, ind := range p.arrangedIndices() {
			s += fmt.Sprintf(fs, p.bars[ind].Name1, p.bars[ind].Name2, p.bars[ind].Name3, p.bars[ind].Ratio, p.bars[ind].Current, p.bars[ind].Total)
		}
		p.writeString(s)
		p.writeString(fmt.Sprintf("\033[%dA", nbars))
	}
	p.writeString(newlines(nbars))
	p.done <- struct{}{}
}
