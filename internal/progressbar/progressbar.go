package progressbar

import (
	"fmt"
	"sync"

	"github.com/fatih/color"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

type ProgressBar interface {
	Wait()
	Bar(string) Bar
	AddBar(string, int) Bar
}

var _ ProgressBar = new(progressBar)

type progressBar struct {
	container *mpb.Progress
	barMap    sync.Map
}

func (s *progressBar) Wait() {
	s.container.Wait()
}

func (s *progressBar) Bar(key string) Bar {
	v, ok := s.barMap.Load(key)
	if ok {
		return v.(*bar)
	}

	return nil
}

func (s *progressBar) AddBar(key string, total int) Bar {
	b := s.container.Add(int64(total),
		mpb.NewBarFiller(mpb.BarStyle().Lbound("[").
			Filler(color.GreenString("=")).
			Tip(color.GreenString(">")).Padding(" ").Rbound("]")),
		mpb.PrependDecorators(
			func() decor.Decorator {
				frames := getSpinner()
				var count uint
				return decor.Any(func(statistics decor.Statistics) string {
					if statistics.Completed {
						return frames[0]
					}
					frame := frames[count%uint(len(frames))]
					count++
					return frame
				})
			}(),
			decor.Name(key, decor.WCSyncWidth),
		),
		mpb.AppendDecorators(
			decor.NewPercentage("%d  "),
			decor.CountersNoUnit("(%d/%d)", decor.WCSyncWidth),
			decor.Any(func(statistics decor.Statistics) string {
				if b := s.Bar(key); b != nil {
					return b.(*bar).getSuffix()
				}
				return ""
			}, decor.WCSyncSpace),
		),
		mpb.BarWidth(15),
	)

	nb := &bar{Bar: b, total: total}
	if total != 0 {
		nb.Add(total)
	}

	s.barMap.Store(key, nb)
	return nb
}

func getSpinner() []string {
	activeState := "[ " + color.GreenString("●") + " ] "
	defaultState := "[   ] "
	return []string{
		activeState,
		activeState,
		activeState,
		defaultState,
		defaultState,
		defaultState,
	}
}

func New() *progressBar {
	return &progressBar{
		container: mpb.New(),
	}
}

type Bar interface {
	Wait()
	TotalInc(delta int)
	TriggerComplete()
	SetSuffix(format string, args ...interface{})
	Incr()
}

type bar struct {
	*mpb.Bar
	sync.WaitGroup
	sync.Mutex
	suffix string
	total  int
}

func (b *bar) TotalInc(delta int) {
	b.Lock()
	b.total += delta
	total := b.total
	b.Unlock()

	b.Add(delta)
	b.Bar.SetTotal(int64(total), false)
}

func (b *bar) TriggerComplete() {
	b.Bar.SetTotal(-1, true)
}

func (b *bar) SetSuffix(format string, args ...interface{}) {
	b.Lock()
	defer b.Unlock()

	b.suffix = fmt.Sprintf(format, args...)
}

func (b *bar) getSuffix() string {
	b.Lock()
	defer b.Unlock()

	return b.suffix
}

func (b *bar) Incr() {
	b.Done()
	b.Bar.Increment()
}
