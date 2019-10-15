package ants

import (
	"errors"
	"time"
)

var (
	// ErrQueueIsFull will be returned when the worker queue is full.
	ErrQueueIsFull = errors.New("the queue is full")

	// ErrQueueLengthIsZero will be returned when trying to insert item to a released worker queue.
	ErrQueueLengthIsZero = errors.New("the queue length is zero")
)

type workerArray interface {
	len() int
	isEmpty() bool
	insert(worker *goWorker) error
	detach() *goWorker
	findOutExpiry(duration time.Duration) []*goWorker
	release()
}

type arrayType int

const (
	stackType arrayType = 1 << iota
	loopQueueType
)

func newWorkerArray(aType arrayType, size int) workerArray {
	switch aType {
	case stackType:
		return newWorkerStack(size)
	case loopQueueType:
		return newWorkerLoopQueue(size)
	default:
		return newWorkerStack(size)
	}
}
