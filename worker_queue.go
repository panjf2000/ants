package ants

import (
	"errors"
	"time"
)

var (
	ErrQueueIsFull       = errors.New("the queue is full")
	ErrQueueLengthIsZero = errors.New("the queue length is zero")
)

type workerQueue interface {
	len() int
	cap() int
	isEmpty() bool
	enqueue(worker *goWorker) error
	dequeue() *goWorker
	releaseExpiry(duration time.Duration) chan *goWorker
	releaseAll()
}

type queueType int

const (
	stackType queueType = 1 << iota
	loopQueueType
)

func newQueue(qType queueType, size int) workerQueue {
	switch qType {
	case stackType:
		return newWorkerStack(size)
	case loopQueueType:
		return newLoopQueue(size)
	default:
		return newWorkerStack(size)
	}
}
