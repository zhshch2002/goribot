package goribot

import (
	"errors"
	"sync"
	"time"
)

type cache struct {
	Data    interface{}
	Expired time.Time
}

type CacheManger struct {
	caches map[string]cache
	lock   sync.Mutex
}

func NewCacheManger() *CacheManger {
	return &CacheManger{
		caches: map[string]cache{},
		lock:   sync.Mutex{},
	}
}

func (cm *CacheManger) Get(k string) (interface{}, bool) {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	res, ok := cm.caches[k]
	if !ok {
		return nil, false
	}
	if res.Expired.Unix() < time.Now().Unix() {
		delete(cm.caches, k)
		return nil, false
	}
	return res.Data, ok
}

func (cm *CacheManger) MustGet(k string) interface{} {
	res, ok := cm.Get(k)
	if !ok {
		panic(errors.New("can't find the cache"))
	}
	return res
}

func (cm *CacheManger) Set(k string, exp time.Duration, updateFunc func() interface{}) interface{} {
	res := updateFunc()
	cm.lock.Lock()
	defer cm.lock.Unlock()
	cm.caches[k] = cache{
		Data:    res,
		Expired: time.Now().Add(exp),
	}
	return res
}

func (cm *CacheManger) GetAndSet(k string, exp time.Duration, updateFunc func() interface{}) interface{} {
	res, ok := cm.Get(k)
	if ok {
		return res
	}
	return cm.Set(k, exp, updateFunc)
}
