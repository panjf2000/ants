package ants

import (
	"errors"
	"time"
)

var (
	// errQueueIsFull will be returned when the worker queue is full.
	errQueueIsFull = errors.New("the queue is full")

	// errQueueIsReleased will be returned when trying to insert item to a released worker queue.
	errQueueIsReleased = errors.New("the queue length is zero")
)

type worker interface {
	run()
	finish()
	when() time.Time
	inputFunc(func())
	inputParam(interface{})
}

type workerQueue interface {
	len() int
	isEmpty() bool
	insert(worker) error
	detach() worker
	staleWorkers(duration time.Duration) []worker
	reset()
}

type queueType int

const (
	queueTypeStack queueType = 1 << iota
	queueTypeLoopQueue
)

func newWorkerArray(qType queueType, size int) workerQueue {
	switch qType {
	case queueTypeStack:
		return newWorkerStack(size)
	case queueTypeLoopQueue:
		return newWorkerLoopQueue(size)
	default:
		return newWorkerStack(size)
	}
}
