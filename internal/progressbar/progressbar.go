package progressbar

import (
	"fmt"
	"sync"

	"github.com/fatih/color"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

type Bar struct {
	sync.WaitGroup
	sync.Mutex
	bar    *mpb.Bar
	suffix string
	total  int
}

func (b *Bar) TotalInc(delta int) {
	if b != nil {
		b.Lock()
		defer b.Unlock()

		b.WaitGroup.Add(delta)
		b.total += delta
		b.bar.SetTotal(int64(b.total), false)
	}
}

func (b *Bar) TriggerComplete() {
	if b != nil {
		b.Lock()
		defer b.Unlock()

		b.bar.SetTotal(int64(b.total), true)
	}
}

func (b *Bar) GetTotal() int {
	if b != nil {
		b.Lock()
		defer b.Unlock()

		return b.total
	}
	return 0
}

func (b *Bar) SetSuffix(format string, args ...interface{}) {
	if b != nil {
		b.Lock()
		defer b.Unlock()

		b.suffix = fmt.Sprintf(format, args...)
	}
}

func (b *Bar) getSuffix() string {
	if b != nil {
		b.Lock()
		defer b.Unlock()

		return b.suffix
	}
	return ""
}

func (b *Bar) Incr() {
	if b != nil {
		b.Lock()
		defer b.Unlock()

		if b.total > 0 {
			b.Done()
		}
		b.bar.Increment()
	}
}

type ProgressBar struct {
	container *mpb.Progress
	barMap    sync.Map
}

func (s *ProgressBar) Wait() {
	if s != nil {
		s.container.Wait()
	}
}

func (s *ProgressBar) DefaultBar() *Bar {
	return s.Bar("")
}

func (s *ProgressBar) Bar(key string) *Bar {
	if s != nil {
		v, ok := s.barMap.Load(key)
		if ok {
			return v.(*Bar)
		}
	}

	return nil
}

func (s *ProgressBar) AddBar(key string, total int) *Bar {
	if s == nil {
		return nil
	}

	bar := s.container.Add(int64(total),
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
				if s != nil {
					if b := s.Bar(key); b != nil {
						return b.getSuffix()
					}
				}
				return ""
			}, decor.WCSyncSpace),
		),
		mpb.BarWidth(15),
	)

	b := &Bar{bar: bar, total: total}
	if total != 0 {
		b.Add(total)
	}
	s.barMap.Store(key, b)
	return b
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

func NewMulti(quiet bool) *ProgressBar {
	var progressBar *ProgressBar
	if quiet {
		return progressBar
	}

	return &ProgressBar{
		container: mpb.New(),
	}
}

func NewSingle(total int, quiet bool) *ProgressBar {
	var progressBar *ProgressBar
	if quiet {
		return progressBar
	}

	progressBar = &ProgressBar{
		container: mpb.New(),
	}

	progressBar.AddBar("", total)
	return progressBar
}
