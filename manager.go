package goribot

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/panjf2000/ants/v2"
	"runtime"
	"time"
)

const ItemsSuffix = "_items"
const TasksSuffix = "_tasks"
const DeduplicateSuffix = "_deduplicate"

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
	s.redis.Del(s.sName + DeduplicateSuffix)
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
				//Log.Info("Waiting for more items")
				time.Sleep(5 * time.Second)
			}
		} else {
			time.Sleep(500 * time.Microsecond)
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

// ReqDeduplicate is an extension can deduplicate new task based on redis to support distributed
func RedisReqDeduplicate(r *redis.Client, sName string) func(s *Spider) {
	return func(s *Spider) {
		s.OnAdd(func(ctx *Context, t *Task) *Task {
			has := GetRequestHash(t.Request)
			res, err := r.SAdd(sName+DeduplicateSuffix, has[:]).Result()
			if err == nil && res == 0 {
				return nil
			}
			return t
		})
	}
}

func RedisDistributed(ro *redis.Options, sName string, useDeduplicate bool, onSeedHandler CtxHandlerFun) func(s *Spider) {
	c1 := redis.NewClient(ro)
	if pong, err := c1.Ping().Result(); pong != "PONG" || err != nil {
		panic("redis connect error " + fmt.Sprint(pong, err))
	}
	var c2 *redis.Client
	if useDeduplicate {
		c2 = redis.NewClient(ro)
		if pong, err := c2.Ping().Result(); pong != "PONG" || err != nil {
			panic("redis connect error " + fmt.Sprint(pong, err))
		}
	}
	return func(s *Spider) {
		s.Scheduler = NewRedisScheduler(c1, sName, 10, onSeedHandler)
		if useDeduplicate {
			s.Use(RedisReqDeduplicate(c2, sName))
		}
		s.AutoStop = false
		s.OnFinish(func(s *Spider) {
			_ = c1.Close()
		})
	}
}
