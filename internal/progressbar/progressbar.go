package progressbar

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

type ProgressBar struct {
	container *mpb.Progress
	bar       *mpb.Bar
	prefix    string
	suffix    string
	total     int
}

func (s *ProgressBar) SetPrefix(format string, args ...interface{}) {
	s.prefix = fmt.Sprintf(format, args...)
}

func (s *ProgressBar) SetTotal(total int, triggerComplete bool) {
	s.total = total
	s.bar.SetTotal(int64(total), triggerComplete)
}

func (s *ProgressBar) SetSuffix(format string, args ...interface{}) {
	s.suffix = fmt.Sprintf(format, args...)
}

func (s *ProgressBar) Incr() {
	s.bar.Increment()
}

func (s *ProgressBar) Wait() {
	s.container.Wait()
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

func New(count int) *ProgressBar {
	var progressBar *ProgressBar
	container := mpb.New()
	bar := container.Add(int64(count),
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
			decor.Any(func(statistics decor.Statistics) string {
				if progressBar != nil {
					return progressBar.prefix
				}
				return ""
			}),
		),
		mpb.AppendDecorators(
			decor.NewPercentage("%d  "),
			decor.Any(func(statistics decor.Statistics) string {
				if progressBar != nil {
					return fmt.Sprintf("(%d/%d) %s", statistics.Current, progressBar.total, progressBar.suffix)
				}
				return ""
			}),
		),
		mpb.BarWidth(15),
	)

	progressBar = &ProgressBar{
		container: container,
		bar:       bar,
		total:     count,
	}

	return progressBar
}
