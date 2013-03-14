package main

import "sync"

type idStore struct {
	locks []bool
	mutex sync.Mutex
}

func createIdStore() *idStore {
	res := &idStore{locks: make([]bool, len(PlayerColors))}
	for idx := range res.locks {
		res.locks[idx] = false
	}
	return res
}

func (ids *idStore) Free(id int) {
	ids.mutex.Lock()
	defer ids.mutex.Unlock()

	ids.locks[id] = false
}

func (ids *idStore) TryGet() int {
	ids.mutex.Lock()
	defer ids.mutex.Unlock()

	for idx, val := range ids.locks {
		if !val {
			ids.locks[idx] = true
			return idx
		}
	}
	return -1
}
