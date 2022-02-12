package progressbar

import "sync"

var _ ProgressBar = new(mockProgressBar)

type mockProgressBar struct {
	sync.Map
}

func NewMock() *mockProgressBar {
	return new(mockProgressBar)
}

func (*mockProgressBar) Wait() {}
func (p *mockProgressBar) Bar(key string) Bar {
	if v, ok := p.Load(key); ok {
		return v.(*mockBar)
	}
	return nil
}
func (p *mockProgressBar) AddBar(key string, total int) Bar {
	b := new(mockBar)
	p.Store(key, b)
	return b
}

var _ Bar = new(mockBar)

type mockBar struct{}

func (*mockBar) Wait()                                        {}
func (*mockBar) TotalInc(delta int)                           {}
func (*mockBar) TriggerComplete()                             {}
func (*mockBar) SetSuffix(format string, args ...interface{}) {}
func (*mockBar) Incr()                                        {}
