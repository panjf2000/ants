package ants

import "errors"

var (
	ErrQueueIsFull       = errors.New("the queue is full")
	ErrQueueLengthIsZero = errors.New("the queue length is zero")
)

type WorkerQueue interface {
	Len() int
	Cap() int
	IsEmpty() bool
	Enqueue(worker interface{}) error
	Dequeue() interface{}
	ReleaseExpiry(isExpiry func(item interface{}) bool) chan interface{}
	ReleaseAll(free func(item interface{}))
}

type QueueType int

const (
	StackType QueueType = 1 << iota
	LoopQueueType
)

func NewQueue(qType QueueType, size int) WorkerQueue {
	switch qType {
	case StackType:
		return NewWorkerStack(size)
	case LoopQueueType:
		return NewLoopQueue(size)
	default:
		return NewWorkerStack(size)
	}
}
