package ants

import (
	"errors"
	"time"
)

// errQueueIsFull will be returned when the worker queue is full.
var errQueueIsFull = errors.New("the queue is full")

type worker interface {
	run()
	finish()
	lastUsedTime() time.Time
	setLastUsedTime(t time.Time)
	inputFunc(func())
	inputArg(any)
}

type workerQueue interface {
	len() int
	isEmpty() bool
	insert(worker) error
	detach() worker
	refresh(duration time.Duration) []worker // clean up the stale workers and return them
	reset()
}

type queueType int

const (
	queueTypeStack queueType = 1 << iota
	queueTypeLoopQueue
)

func newWorkerQueue(qType queueType, size int) workerQueue {
	switch qType {
	case queueTypeStack:
		return newWorkerStack(size)
	case queueTypeLoopQueue:
		return newWorkerLoopQueue(size)
	default:
		return newWorkerStack(size)
	}
}
