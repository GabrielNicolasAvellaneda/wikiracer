package control

import "sync"

// simple thread safe map.
type visitedMap struct {
	sync.Mutex

	m      map[string]bool
	length uint64
}

func (v *visitedMap) visited(id string) bool {
	v.Lock()
	defer v.Unlock()

	_, ok := v.m[id]
	if ok {
		return true
	}
	v.m[id] = true
	v.length++
	return false
}

func (v *visitedMap) len() uint64 {
	v.Lock()
	l := v.length
	v.Unlock()
	return l
}
