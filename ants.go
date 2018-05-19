package ants

import "math"

const DEFAULT_POOL_SIZE = math.MaxInt32

var defaultPool = NewPool(DEFAULT_POOL_SIZE)

func Push(task f) error {
	return defaultPool.Push(task)
}

func Size() int {
	return int(defaultPool.Running())
}

func Cap() int {
	return int(defaultPool.Cap())
}
