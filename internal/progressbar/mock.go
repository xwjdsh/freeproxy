package progressbar

var _ ProgressBar = new(mockProgressBar)

type mockProgressBar struct{}

func NewMock() *mockProgressBar {
	return new(mockProgressBar)
}

func (*mockProgressBar) Wait()                  {}
func (*mockProgressBar) DefaultBar() Bar        { return new(mockBar) }
func (*mockProgressBar) Bar(string) Bar         { return new(mockBar) }
func (*mockProgressBar) AddBar(string, int) Bar { return new(mockBar) }

var _ Bar = new(mockBar)

type mockBar struct{}

func (*mockBar) Wait()                                        {}
func (*mockBar) TotalInc(delta int)                           {}
func (*mockBar) TriggerComplete()                             {}
func (*mockBar) SetSuffix(format string, args ...interface{}) {}
func (*mockBar) Incr()                                        {}
