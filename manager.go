package goribot

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/go-redis/redis"
	"github.com/panjf2000/ants/v2"
	"runtime"
)

const ItemsSuffix = "_items"
const TasksSuffix = "_tasks"

type item struct {
	Data interface{}
}

type Manager struct {
	itemPool       *ants.Pool
	redis          *redis.Client
	sName          string
	onItemHandlers []func(i interface{}) interface{}
}

func NewManager(redis *redis.Client, sName string) *Manager {
	ip, err := ants.NewPool(runtime.NumCPU())
	if err != nil {
		panic(err)
	}
	return &Manager{
		itemPool:       ip,
		redis:          redis,
		sName:          sName,
		onItemHandlers: []func(i interface{}) interface{}{},
	}
}

func (s *Manager) OnItem(fn func(i interface{}) interface{}) {
	s.onItemHandlers = append(s.onItemHandlers, fn)
}

func (s *Manager) handleOnItem(i interface{}) {
	for _, fn := range s.onItemHandlers {
		i = fn(i)
		if i == nil {
			return
		}
	}
}

func (s *Manager) SetItemPoolSize(i int) {
	s.itemPool.Tune(i)
}

func (s *Manager) Run() {
	for {
		if s.itemPool.Free() > 0 {
			if i := s.GetItem(); i != nil {
				err := s.itemPool.Submit(func() {
					s.handleOnItem(i)
				})
				if errors.Is(err, ants.ErrPoolClosed) {
					panic(ErrRunFinishedSpider)
				}
			} else if s.itemPool.Running() == 0 {
				break
			}
		}
		runtime.Gosched()
	}
}

func (s *Manager) GetItem() interface{} {
	res, err := s.redis.LPop(s.sName + ItemsSuffix).Bytes()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			Log.Error(err)
		}
		return nil
	}
	dec := gob.NewDecoder(bytes.NewReader(res))
	item := item{}
	err = dec.Decode(&item)
	if err != nil {
		Log.Error(err)
	}
	return item.Data
}

func (s *Manager) SendReq(req *Request) {
	var buffer bytes.Buffer
	ecoder := gob.NewEncoder(&buffer)
	err := ecoder.Encode(req)
	if err != nil {
		Log.Error(err)
		return
	}
	err = s.redis.LPush(s.sName+TasksSuffix, buffer.Bytes()).Err()
	if err != nil {
		Log.Error(err)
	}
}

// Scheduler is default scheduler of goribot
type RedisScheduler struct {
	redis     *redis.Client
	sName     string
	fn        []CtxHandlerFun
	batchSize int
	base      *BaseScheduler
}

func NewRedisScheduler(redis *redis.Client, sName string, bs int, fn ...CtxHandlerFun) *RedisScheduler {
	return &RedisScheduler{redis, sName, fn, bs, NewBaseScheduler(false)}
}
func (s *RedisScheduler) loadRedisTask() {
	i := 0
	for i < s.batchSize {
		res, err := s.redis.LPop(s.sName + TasksSuffix).Bytes()
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				Log.Error(err)
			}
			return
		}
		dec := gob.NewDecoder(bytes.NewReader(res))
		req := &Request{}
		err = dec.Decode(req)
		s.base.AddTask(NewTask(req, s.fn...))
		i += 1
	}
}

func (s *RedisScheduler) GetTask() *Task {
	t := s.base.GetTask()
	if t == nil {
		s.loadRedisTask()
		t = s.base.GetTask()
	}
	return t

}
func (s *RedisScheduler) GetItem() interface{} {
	return s.base.GetItem()
}
func (s *RedisScheduler) AddTask(t *Task) {
	s.base.AddTask(t)
}
func (s *RedisScheduler) AddItem(i interface{}) {
	s.base.AddItem(i)
	var buffer bytes.Buffer
	ecoder := gob.NewEncoder(&buffer)
	err := ecoder.Encode(item{Data: i})
	if err != nil {
		Log.Error(err)
		return
	}
	err = s.redis.LPush(s.sName+ItemsSuffix, buffer.Bytes()).Err()
	if err != nil {
		Log.Error(err)
		return
	}
}
func (s *RedisScheduler) IsTaskEmpty() bool {
	s.loadRedisTask()
	return s.base.IsItemEmpty()
}
func (s *RedisScheduler) IsItemEmpty() bool {
	l, err := s.redis.LLen(s.sName + ItemsSuffix).Result()
	return l == 0 || err != nil
}
