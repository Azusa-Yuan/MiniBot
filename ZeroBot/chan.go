package zero

import (
	"sync"
	"time"
)

type eventChan struct {
	sync.Mutex
	c chan eventItem
}

type eventItem struct {
	response []byte
	caller   APICaller
}

func newchan(chansize uint) eventChan {
	return eventChan{
		c: make(chan eventItem, chansize),
	}
}

func (evc *eventChan) processEvent(response []byte, caller APICaller) {
	select {
	case evc.c <- eventItem{
		response: response,
		caller:   caller,
	}:
		return
	// 阻塞了则丢弃旧消息，增加最新消息
	default:
		select {
		case <-evc.c:
			evc.c <- eventItem{
				response: response,
				caller:   caller,
			}
		default:
			return
		}
	}
}

// loop 循环处理事件
//
//	latency 延迟 latency 再处理事件
func (evc *eventChan) loop(latency, maxwait time.Duration, process func([]byte, APICaller, time.Duration)) {
	go func() {
		timer := time.NewTicker(latency)
		if latency < time.Millisecond {
			latency = time.Millisecond
		}
		for range timer.C {
			item := <-evc.c
			process(item.response, item.caller, maxwait)
		}
	}()
}
