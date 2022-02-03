package counter

import "sync/atomic"

type Count int32

func (c *Count) Inc() int {
	return int(atomic.AddInt32((*int32)(c), 1))
}

func (c *Count) Get() int {
	return int(atomic.LoadInt32((*int32)(c)))
}
