package ants

import "time"

type WorkerQueue interface {
	Len() int
	Cap() int
	IsEmpty() bool
	Enqueue(worker *goWorker) error
	Dequeue() *goWorker
	ReleaseExpiry(expiry time.Duration) chan *goWorker
	ReleaseAll()
}

type QueueType int

const (
	ArrayQueueType 	QueueType = 1 << iota
	LoopQueueType
)

func NewQueue(qType QueueType, size int) WorkerQueue {
	switch qType {
	case ArrayQueueType: 	return NewSliceQueue(size)
	case LoopQueueType:		return NewLoopQueue(size)
	default: 				return NewSliceQueue(size)
	}
}