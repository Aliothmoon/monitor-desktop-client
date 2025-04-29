package utils

import (
	"log"
	"sync"
	"time"
)

func Go(f func()) {
	go func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
		f()
	}()
}

type AdvancedThrottler struct {
	duration  time.Duration
	trigger   chan struct{}
	execMutex sync.Mutex
}

func NewAdvancedThrottler(d time.Duration) *AdvancedThrottler {
	t := &AdvancedThrottler{
		duration: d,
		trigger:  make(chan struct{}, 2),
	}
	Go(t.schedule)
	return t
}

// Do 外部调用入口
func (t *AdvancedThrottler) Do(fn func()) {
	select {
	case t.trigger <- struct{}{}:
		t.execMutex.Lock()
		defer t.execMutex.Unlock()
		fn()
	default:
		// 冷却期内跳过执行
	}
}

// 定时重置通道
func (t *AdvancedThrottler) schedule() {
	for range time.Tick(t.duration) {
		select {
		case <-t.trigger:
		default:
		}
	}
}
