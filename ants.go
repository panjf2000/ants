package ants

const DEFAULT_POOL_SIZE = 1000

var defaultPool = NewPool(DEFAULT_POOL_SIZE)

func Push(task f) error {
	return defaultPool.Push(task)
}

func Running() int {
	return defaultPool.Running()
}

func Cap() int {
	return defaultPool.Cap()
}

func Free() int {
	return defaultPool.Free()
}
