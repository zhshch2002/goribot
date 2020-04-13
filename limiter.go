package goribot

import (
	"github.com/gobwas/glob"
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
		s.Downloader.OnReq(func(req *Request) *Request {
			for k, r := range rules {
				if r.Match(req.URL) {
					if r.Delay > 0 {
						rules[k].delayLock.Lock()
						since := time.Since(r.lastReqTime)
						if since < r.Delay {
							time.Sleep(r.Delay - since)
						}
						rules[k].lastReqTime = time.Now()
						rules[k].delayLock.Unlock()
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
					}
					return req
				}
			}
			return req
		})
		s.Downloader.OnResp(func(resp *Response) *Response {
			for k, r := range rules {
				if r.Match(resp.Request.URL) {
					if r.Parallelism > 0 {
						atomic.AddInt64(&rules[k].workingParallelism, -1)
					}
					return resp
				}
			}
			return resp
		})
		s.OnAdd(func(ctx *Context, t *Task) *Task {
			for _, r := range rules {
				if r.Match(t.Request.URL) {
					if r.Allow == Allow {
						return t
					} else {
						return nil
					}
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
