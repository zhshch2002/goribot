package goribot

import (
	"github.com/gobwas/glob"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type LimitRuleAllow uint8

const (
	NotSet LimitRuleAllow = iota
	Allow
	Disallow
)

type LimitRule struct {
	Regexp, Glob       string
	Allow              LimitRuleAllow
	Parallelism        int64
	workingParallelism int64
	Rate               int64
	rateLeft           int64
	Delay              time.Duration
	RandomDelay        time.Duration
	MaxReq             int64
	reqLeft            int64
	MaxDepth           int64
	lastReqTime        time.Time
	compiledRegexp     *regexp.Regexp
	compiledGlob       glob.Glob
	delayLock          sync.Mutex
}

func (s *LimitRule) Match(u *url.URL) bool {
	match := false
	if s.compiledGlob != nil {
		match = s.compiledGlob.Match(strings.ToLower(u.Host))
	} else {
		match = s.compiledRegexp.MatchString(strings.ToLower(u.Host))
	}
	return match
}

func Limiter(WhiteList bool, rules ...*LimitRule) func(s *Spider) {
	for k, r := range rules {
		if r.Allow == NotSet {
			rules[k].Allow = Allow
		}
		rules[k].rateLeft = r.Rate
		rules[k].reqLeft = r.MaxReq
		rules[k].delayLock = sync.Mutex{}
		if rules[k].Glob != "" {
			rules[k].compiledGlob = glob.MustCompile(rules[k].Glob)
		} else {
			rules[k].compiledRegexp = regexp.MustCompile(rules[k].Regexp)
		}
	}
	rateCtl := true
	go func() {
		for rateCtl {
			time.Sleep(1 * time.Second)
			for k, _ := range rules {
				atomic.StoreInt64(&rules[k].rateLeft, rules[k].Rate)
			}
		}
	}()
	return func(s *Spider) {
		s.Downloader.AddMiddleware(func(req *Request, next func(req *Request) (resp *Response, err error)) (resp *Response, err error) {
			for k, r := range rules {
				if r.Match(req.URL) {
					if r.Delay > 0 || r.RandomDelay > 0 {
						rules[k].delayLock.Lock()
						since := time.Since(r.lastReqTime)
						if since < r.Delay {
							time.Sleep(r.Delay - since)
						}
						if r.RandomDelay > 0 {
							ra := rand.New(rand.NewSource(time.Now().Unix()))
							time.Sleep(time.Duration(ra.Int63n(int64(r.RandomDelay))))
						}
						rules[k].lastReqTime = time.Now()
						rules[k].delayLock.Unlock()
						return next(req)
					} else if r.Rate > 0 {
						wait := true
						for wait {
							if atomic.LoadInt64(&rules[k].rateLeft) > 0 {
								atomic.AddInt64(&rules[k].rateLeft, -1)
								wait = false
							} else {
								time.Sleep(500 * time.Microsecond)
							}
						}
						return next(req)
					} else if r.Parallelism > 0 {
						wait := true
						for wait {
							if atomic.LoadInt64(&rules[k].workingParallelism) < r.Parallelism {
								atomic.AddInt64(&rules[k].workingParallelism, 1)
								wait = false
							} else {
								time.Sleep(500 * time.Microsecond)
							}
						}
						resp, err := next(req)
						atomic.AddInt64(&rules[k].workingParallelism, -1)
						return resp, err
					} else {
						return next(req)
					}
				}
			}
			return next(req)
		})
		s.OnAdd(func(ctx *Context, t *Task) *Task {
			for k, r := range rules {
				if r.Match(t.Request.URL) {
					if r.Allow == Disallow {
						return nil
					}
					if r.MaxDepth > 0 {
						//fmt.Println(t.Request.Depth)
						if int64(t.Request.Depth) > r.MaxDepth {
							return nil
						}
					}
					if r.MaxReq > 0 {
						if atomic.LoadInt64(&rules[k].reqLeft) > 0 {
							atomic.AddInt64(&rules[k].reqLeft, -1)
						} else {
							return nil
						}
					}
					return t
				}
			}
			if WhiteList {
				return nil
			} else {
				return t
			}
		})
		s.OnFinish(func(s *Spider) {
			rateCtl = false
		})
	}
}
