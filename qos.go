package ants

import (
	"sync"
	"sync/atomic"
	"time"
)

const DefaultQosDuration = 1 * time.Second

type Qos struct {
	qosDuration  time.Duration // Duration for QoS (Quality of Service) interval
	qosLimit     int32         // Maximum number of QoS executions allowed within the interval
	qosExecuted  int32         // Number of QoS executions that have occurred
	qosResetter  *time.Ticker  // Ticker for resetting QoS
	qosResetCond sync.Cond     // Condition for waiting for QoS reset
}

// newQos Creates a new Qos instance with the specified duration and limit.
func newQos(qosDuration time.Duration, qosLimit int) *Qos {
	q := &Qos{
		qosLimit:     int32(qosLimit),
		qosResetCond: sync.Cond{L: &sync.RWMutex{}},
	}
	q.SetQosDuration(qosDuration)
	return q
}

// QosResetterAsync Starts an asynchronous QoS resetter.
func (q *Qos) QosResetterAsync() {
	if q.qosResetter != nil {
		return // Resetter already running
	}
	q.qosResetter = time.NewTicker(q.qosDuration)
	go func() {
		for range q.qosResetter.C {
			atomic.StoreInt32(&q.qosExecuted, 0)
			q.qosResetCond.Broadcast()
		}
	}()
}

// WaitQosUnlock Waits until the QoS unlock condition is met.
func (q *Qos) WaitQosUnlock() {
	if q.QosCheck() {
		return
	}
	q.qosResetCond.L.Lock()
	for !q.QosCheck() {
		q.qosResetCond.Wait()
	}
	q.qosResetCond.L.Unlock()
}

// QosCheck Checks if the QoS limit has been reached.
func (q *Qos) QosCheck() bool {
	return q.qosLimit == 0 || atomic.LoadInt32(&q.qosExecuted) < q.qosLimit
}

func (q *Qos) QosLimit() int32 {
	return q.qosLimit
}

func (q *Qos) SetQosLimit(limit int32) {
	q.qosLimit = limit
}

func (q *Qos) QosDuration() time.Duration {
	return q.qosDuration
}

func (q *Qos) SetQosDuration(duration time.Duration) {
	if duration == 0 {
		q.qosDuration = DefaultQosDuration
	} else {
		q.qosDuration = duration
	}
}

func (q *Qos) AddQosExecuted() {
	if q.qosLimit == 0 {
		return
	}
	atomic.AddInt32(&q.qosExecuted, 1)
}
