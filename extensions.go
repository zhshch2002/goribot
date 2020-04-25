package goribot

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/op/go-logging"
	"github.com/slyrz/robots"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type JsonItem struct {
	Data interface{}
}

// SaveItemsAsCSV is a extension save items to a json file
func SaveItemsAsJSON(f *os.File) func(s *Spider) {
	lock := sync.Mutex{}
	f.WriteString("[")
	gotFrist := false
	return func(s *Spider) {
		s.OnItem(func(i interface{}) interface{} {
			if data, ok := i.(JsonItem); ok {
				lock.Lock()
				defer lock.Unlock()
				if gotFrist {
					_, err := f.WriteString(",")
					if err != nil {
						Log.Error(err)
					}
				} else {
					gotFrist = true
				}
				res, err := json.Marshal(data.Data)
				if err != nil {
					Log.Error(err)
				}
				_, err = f.Write(res)
				if err != nil {
					Log.Error(err)
				}

			}
			return i
		})
		s.OnFinish(func(s *Spider) {
			_, err := f.WriteString("]")
			if err != nil {
				Log.Error(err)
			}
		})
	}
}

type CsvItem []string

// SaveItemsAsCSV is a extension save items to a csv file
func SaveItemsAsCSV(f *os.File) func(s *Spider) {
	lock := sync.Mutex{}
	w := csv.NewWriter(f)
	return func(s *Spider) {
		s.OnItem(func(i interface{}) interface{} {
			if data, ok := i.(CsvItem); ok {
				lock.Lock()
				defer lock.Unlock()
				err := w.Write(data)
				if err != nil {
					Log.Error(err)
				}
			}
			return i
		})
		s.OnFinish(func(s *Spider) {
			w.Flush()
		})
	}
}

// Retry is a extension make a new request when get response with error
func Retry(maxTimes int, okcode ...int) func(s *Spider) {
	return func(s *Spider) {
		s.OnError(func(ctx *Context, err error) {
			if e, ok := err.(DownloaderErr); ok {
				if e.Request != nil {
					if t, ok := e.Request.Meta["RetryTimes"]; !ok || t.(int) < maxTimes {
						req := e.Request
						if !ok {
							req.Meta["RetryTimes"] = 1
						} else {
							req.Meta["RetryTimes"] = req.Meta["RetryTimes"].(int) + 1
						}
						Log.Info("Request to", req.URL, "[tried", req.Meta["RetryTimes"], "times]", "got error.Retry.")
						s.AddTask(req, ctx.Handlers...)
					}
				}
			}
		})
		if len(okcode) > 0 {
			s.OnResp(func(ctx *Context) {
				for _, c := range okcode {
					if ctx.Resp.StatusCode == c {
						return
					}
					if t, ok := ctx.Req.Meta["RetryTimes"]; !ok || t.(int) < maxTimes {
						req := ctx.Req
						if !ok {
							req.Meta["RetryTimes"] = 1
						} else {
							req.Meta["RetryTimes"] = req.Meta["RetryTimes"].(int) + 1
						}
						Log.Info("Request to", req.URL, "[tried", req.Meta["RetryTimes"], "times]", "got error.Retry.")
						s.AddTask(req, ctx.Handlers...)
						ctx.Abort()
					}
				}
			})
		}
	}
}

// RobotsTxt is an extension can parse the robots.txt and follow it
func RobotsTxt(baseUrl, ua string) func(s *Spider) {
	if !strings.HasSuffix(baseUrl, "/") {
		baseUrl += "/"
	}
	resp, err := NewBaseDownloader().Do(GetReq(baseUrl + "robots.txt"))
	if err != nil {
		Log.Error("get robots.txt error", err)
		return func(s *Spider) {}
	}

	RobotsTxt := robots.New(strings.NewReader(resp.Text), ua)
	return func(s *Spider) {
		s.OnReq(func(ctx *Context, req *Request) *Request {
			if RobotsTxt.Allow(req.URL.Path) {
				return req
			}
			return nil

		})
	}
}

var spiderLogErrorFormat = logging.MustStringFormatter(
	`%{time:15:04:05.000} %{shortfile} ▶ %{level:.4s} %{id:03x} %{message}`,
)

type ErrorItem struct {
	Ctx *Context
	Msg string
}

// SpiderLogError is a extension logs special or error response
func SpiderLogError(f *os.File) func(s *Spider) {
	log := logging.MustGetLogger("goribot")
	backend := logging.NewBackendFormatter(logging.NewLogBackend(f, "", 0), spiderLogErrorFormat)
	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(logging.DEBUG, "")
	log.SetBackend(backendLeveled)
	return func(s *Spider) {
		s.OnError(func(ctx *Context, err error) {
			log.Error(
				"\n",
				"Got 'OnError' with url ", ctx.Req.URL, "\n",
				"Err:", err, "\n",
				string(debug.Stack()), "\n",
				"Req Header:", ctx.Req.Header, "Proxy:", ctx.Req.ProxyURL, "Err:", ctx.Req.Err, "\n",
				func() string {
					if ctx.Resp != nil {
						return fmt.Sprint(
							"Resp Header:", ctx.Resp.Header,
							" Len:", len(ctx.Resp.Body),
							"\n",
							" Text:", ctx.Resp.Text,
						)
					}
					return ""
				}(), "\n",
			)
		})
		s.OnItem(func(i interface{}) interface{} {
			if e, ok := i.(ErrorItem); ok {
				ctx, msg := e.Ctx, e.Msg
				log.Error(
					"\n",
					"Got 'ErrorItem' with url ", ctx.Req.URL, "\n",
					"Msg:", msg, "\n",
					"Req Header:", ctx.Req.Header, "Proxy:", ctx.Req.ProxyURL, "Err:", ctx.Req.Err, "\n",
					func() string {
						if ctx.Resp != nil {
							return fmt.Sprint(
								"Resp Status code:", ctx.Resp.StatusCode, " Header:", ctx.Resp.Header,
								" Len:", len(ctx.Resp.Body),
								" Text:", ctx.Resp.Text,
							)
						}
						return ""
					}(), "\n",
				)
			}
			return i
		})
	}
}

// SpiderLogPrint is a extension print spider working status
func SpiderLogPrint() func(s *Spider) {
	var t, i int64 = 0, 0
	const n int64 = 5 // 打印时间间隔
	go func() {
		for {
			tt, ii := atomic.LoadInt64(&t), atomic.LoadInt64(&i)
			Log.Info(
				"Handled", tt, "tasks and", ii, "items", "in", n, "sec",
				tt/n, "task/sec", ii/n, "item/sec",
			)
			atomic.SwapInt64(&t, 0)
			atomic.SwapInt64(&i, 0)
			time.Sleep(time.Duration(n) * time.Second)
		}
	}()
	return func(s *Spider) {
		s.OnStart(func(s *Spider) {
			Log.Info("Spider start")
		})
		s.OnReq(func(ctx *Context, req *Request) *Request {
			atomic.AddInt64(&t, 1)
			return req
		})
		s.OnItem(func(item interface{}) interface{} {
			atomic.AddInt64(&i, 1)
			return item
		})
		s.OnFinish(func(s *Spider) {
			Log.Info("Spider finish")
		})
	}
}

// RefererFiller is an extension can add Referer for new task
func RefererFiller() func(s *Spider) {
	return func(s *Spider) {
		s.OnAdd(func(ctx *Context, t *Task) *Task {
			if ctx != nil {
				t.Request.Header.Set("Referer", ctx.Resp.Request.URL.String())
			}
			return t
		})
	}
}

// SetDepthFirst is an extension change Scheduler DepthFirst setting
func SetDepthFirst(d bool) func(s *Spider) {
	return func(s *Spider) {
		if _, ok := s.Scheduler.(*BaseScheduler); !ok {
			panic("spider is not using BaseScheduler from goribot")
		}
		s.Scheduler = NewBaseScheduler(d)
	}
}

// AddCookieToJar is an extension add a cookie to downloader's cookie jar
func AddCookieToJar(urlAddr string, cookies ...*http.Cookie) func(s *Spider) {
	return func(s *Spider) {
		if d, ok := s.Downloader.(*BaseDownloader); ok {
			u, err := url.Parse(urlAddr)
			if err != nil {
				Log.Error(fmt.Errorf("add cookie to jar fail %w", err))
			}
			d.Client.Jar.SetCookies(u, cookies)
		} else {
			panic("spider is not using BaseDownloader from goribot")
		}

	}
}

// ReqDeduplicate is an extension can deduplicate new task
func ReqDeduplicate() func(s *Spider) {
	CrawledHash := map[[md5.Size]byte]struct{}{}
	lock := sync.Mutex{}
	return func(s *Spider) {
		s.OnAdd(func(ctx *Context, t *Task) *Task {
			if _, ok := t.Request.Meta["RetryTimes"]; ok {
				return t
			}
			has := GetRequestHash(t.Request)

			lock.Lock()
			defer lock.Unlock()

			if _, ok := CrawledHash[has]; ok {
				return nil
			}

			CrawledHash[has] = struct{}{}
			return t
		})
	}
}

// RandomUserAgent is an extension can set random proxy url for new task
func RandomProxy(p ...string) func(s *Spider) {
	var RandSrc int64
	return func(s *Spider) {
		s.OnReq(func(ctx *Context, req *Request) *Request {
			_, ok := req.Meta["RandomProxy"]
			if req.ProxyURL == "" || ok {
				RandSrc += 1
				ra := rand.New(rand.NewSource(time.Now().Unix() + RandSrc))
				RandSrc = ra.Int63()
				req.ProxyURL = p[ra.Intn(len(p))]
				req.Meta["RandomProxy"] = struct{}{}
			}
			return req
		})
	}
}

// RandomUserAgent is an extension can set random User-Agent for new task
func RandomUserAgent() func(s *Spider) {
	var RandSrc int64
	return func(s *Spider) {
		s.OnReq(func(ctx *Context, req *Request) *Request {
			_, ok := req.Meta["RandomUserAgent"]
			if req.Request.Header.Get("User-Agent") == "" || ok {
				RandSrc += 1
				rs := rand.NewSource(time.Now().Unix() + RandSrc)
				ra := rand.New(rs)
				RandSrc = ra.Int63()
				req.Request.Header.Set("User-Agent", uaList[ra.Intn(len(uaList))])
				req.Meta["RandomUserAgent"] = struct{}{}
			}
			return req
		})
	}
}

var uaList = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML like Gecko) Chrome/44.0.2403.155 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2227.1 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2227.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2227.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.3; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2226.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.4; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2225.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.3; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2225.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2224.3 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.93 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.2062.124 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.2049.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 4.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.2049.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.67 Safari/537.36",
	"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.67 Safari/537.36",
	"Mozilla/5.0 (X11; OpenBSD i386) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.125 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1944.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/35.0.3319.102 Safari/537.36",
	"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/35.0.2309.372 Safari/537.36",
	"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/35.0.2117.157 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/35.0.1916.47 Safari/537.36",
	"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/34.0.1866.237 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/34.0.1847.137 Safari/4E423F",
	"Mozilla/5.0 (X11; Linux i686; rv:64.0) Gecko/20100101 Firefox/64.0",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:64.0) Gecko/20100101 Firefox/64.0",
	"Mozilla/5.0 (X11; Linux i586; rv:63.0) Gecko/20100101 Firefox/63.0",
	"Mozilla/5.0 (Windows NT 6.2; WOW64; rv:63.0) Gecko/20100101 Firefox/63.0",
	"Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10.10; rv:62.0) Gecko/20100101 Firefox/62.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:10.0) Gecko/20100101 Firefox/62.0",
	"Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10.13; ko; rv:1.9.1b2) Gecko/20081201 Firefox/60.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Firefox/58.0.1",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:54.0) Gecko/20100101 Firefox/58.0",
	"Mozilla/5.0 (Windows NT 6.3; WOW64; rv:52.59.12) Gecko/20160044 Firefox/52.59.12",
	"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9a1) Gecko/20060814 Firefox/51.0",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:46.0) Gecko/20120121 Firefox/46.0",
	"Mozilla/5.0 (Windows NT 10.0; WOW64; rv:45.66.18) Gecko/20177177 Firefox/45.66.18",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:40.0) Gecko/20100101 Firefox/40.1",
	"Mozilla/5.0 (Windows NT 6.3; rv:36.0) Gecko/20100101 Firefox/36.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10; rv:33.0) Gecko/20100101 Firefox/33.0",
	"Mozilla/5.0 (X11; Linux i586; rv:31.0) Gecko/20100101 Firefox/31.0",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:31.0) Gecko/20130401 Firefox/31.0",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:28.0) Gecko/20100101 Firefox/31.0",
	"Mozilla/5.0 (Windows NT 5.1; rv:31.0) Gecko/20100101 Firefox/31.0",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:29.0) Gecko/20120101 Firefox/29.0",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:25.0) Gecko/20100101 Firefox/29.0",
	"Mozilla/5.0 (X11; OpenBSD amd64; rv:28.0) Gecko/20100101 Firefox/28.0",
	"Mozilla/5.0 (X11; Linux x86_64; rv:28.0) Gecko/20100101 Firefox/28.0",
	"Mozilla/5.0 (Windows NT 6.1; rv:27.3) Gecko/20130101 Firefox/27.3",
	"Mozilla/5.0 (Windows NT 6.2; Win64; x64; rv:27.0) Gecko/20121011 Firefox/27.0",
	"Mozilla/5.0 (Windows NT 6.2; rv:20.0) Gecko/20121202 Firefox/26.0",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:25.0) Gecko/20100101 Firefox/25.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.6; rv:25.0) Gecko/20100101 Firefox/25.0",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:24.0) Gecko/20100101 Firefox/24.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML like Gecko) Chrome/51.0.2704.79 Safari/537.36 Edge/14.14931",
	"Chrome (AppleWebKit/537.1; Chrome50.0; Windows NT 6.3) AppleWebKit/537.36 (KHTML like Gecko) Chrome/51.0.2704.79 Safari/537.36 Edge/14.14393",
	"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML like Gecko) Chrome/46.0.2486.0 Safari/537.36 Edge/13.9200",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML like Gecko) Chrome/46.0.2486.0 Safari/537.36 Edge/13.10586",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.246",
	"Mozilla/5.0 (Windows; U; Windows NT 6.1; rv:2.2) Gecko/20110201",
	"Mozilla/5.0 (Windows; U; Windows NT 6.1; it; rv:2.0b4) Gecko/20100818",
	"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9a3pre) Gecko/20070330",
	"Mozilla/5.0 (Windows; U; Windows NT 5.0; en-US; rv:1.9.2a1pre) Gecko",
	"Mozilla/5.0 (Windows; U; Windows NT 5.1; pl; rv:1.9.2.3) Gecko/20100401 Lightningquail/3.6.3",
	"Mozilla/5.0 (X11; ; Linux i686; rv:1.9.2.20) Gecko/20110805",
	"Mozilla/5.0 (Windows; U; Windows NT 5.1; fr; rv:1.9.2.13) Gecko/20101203 iPhone",
	"Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10.6; en-US; rv:1.9.2.13; ) Gecko/20101203",
	"Mozilla/5.0 (Windows; U; Windows NT 6.0; en-US; rv:1.9.1b3) Gecko/20090305",
	"Mozilla/5.0 (Windows; U; Windows NT 5.1; zh-TW; rv:1.9.0.9) Gecko/2009040821",
	"Mozilla/5.0 (X11; U; Linux i686; ru; rv:1.9.0.8) Gecko/2009032711",
	"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.7) Gecko/2009032803",
	"Mozilla/5.0 (Windows; U; Windows NT 6.0; en-GB; rv:1.9.0.7) Gecko/2009021910 MEGAUPLOAD 1.0",
	"Mozilla/5.0 (Windows; U; BeOS; en-US; rv:1.9.0.7) Gecko/2009021910",
	"Mozilla/5.0 (X11; U; Linux i686; pl-PL; rv:1.9.0.6) Gecko/2009020911",
	"Mozilla/5.0 (X11; U; Linux i686; en; rv:1.9.0.6) Gecko/20080528",
	"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.6) Gecko/2009020409",
	"Mozilla/5.0 (X11; U; Linux x86_64; en-US; rv:1.9.0.3) Gecko/2008092814 (Debian-3.0.1-1)",
	"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.3) Gecko/2008092816",
	"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.3) Gecko/2008090713",
	"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.2) Gecko Fedora/1.9.0.2-1.fc9",
	"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.14) Gecko/2009091010",
	"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.10) Gecko/2009042523",
	"Mozilla/5.0 (X11; U; Linux x86_64; en-US; rv:1.9.0.1) Gecko/2008072610",
	"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.1) Gecko/2008072820 Ubuntu/8.04 (hardy) (Linux Mint)",
	"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.1) Gecko",
	"Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10.5; en-US; rv:1.9.0.1) Gecko/2008070206",
	"Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10.5; en-au; rv:1.9.0.1) Gecko/2008070206",
	"Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9) Gecko",
	"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.9) Gecko",
	"Mozilla/5.0 (Windows; U; Windows NT 5.1; cs; rv:1.9) Gecko/2008052906",
	"Mozilla/5.0 (Windows; U; Windows NT 5.0; en-US; rv:1.8b2) Gecko/20050702",
	"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8b) Gecko/20050217",
	"Mozilla/5.0 (Windows; U; Windows NT 5.1; de-AT; rv:1.8b) Gecko/20050217",
	"Mozilla/5.0 (Macintosh; U; PPC Mac OS X Mach-O; en-US; rv:1.8b) Gecko/20050217",
	"Mozilla/5.0 (Windows; U; Win98; en-US; rv:1.8a6) Gecko/20050111",
	"Mozilla/5.0 (Windows; U; Windows NT 5.1; de-AT; rv:1.8a5) Gecko/20041122",
	"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8a4) Gecko/20040927",
	"Mozilla/5.0 (Windows; U; Windows NT 5.0; de-AT; rv:1.8a4) Gecko/20040927",
	"Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US; rv:1.8a3) Gecko/20040817",
	"Opera/9.80 (X11; Linux i686; Ubuntu/14.10) Presto/2.12.388 Version/12.16",
	"Opera/9.80 (Macintosh; Intel Mac OS X 10.14.1) Presto/2.12.388 Version/12.16",
	"Opera/9.80 (Windows NT 6.0) Presto/2.12.388 Version/12.14",
	"Mozilla/5.0 (Windows NT 6.0; rv:2.0) Gecko/20100101 Firefox/4.0 Opera 12.14",
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.0) Opera 12.14",
	"Opera/12.80 (Windows NT 5.1; U; en) Presto/2.10.289 Version/12.02",
	"Opera/9.80 (Windows NT 6.1; U; es-ES) Presto/2.9.181 Version/12.00",
	"Opera/9.80 (Windows NT 5.1; U; zh-sg) Presto/2.9.181 Version/12.00",
	"Opera/12.0(Windows NT 5.2;U;en)Presto/22.9.168 Version/12.00",
	"Opera/12.0(Windows NT 5.1;U;en)Presto/22.9.168 Version/12.00",
	"Mozilla/5.0 (Windows NT 5.1) Gecko/20100101 Firefox/14.0 Opera/12.0",
	"Opera/9.80 (Windows NT 6.1; WOW64; U; pt) Presto/2.10.229 Version/11.62",
	"Opera/9.80 (Windows NT 6.0; U; pl) Presto/2.10.229 Version/11.62",
	"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; fr) Presto/2.9.168 Version/11.52",
	"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; de) Presto/2.9.168 Version/11.52",
	"Opera/9.80 (Windows NT 5.1; U; en) Presto/2.9.168 Version/11.51",
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; de) Opera 11.51",
	"Opera/9.80 (X11; Linux x86_64; U; fr) Presto/2.9.168 Version/11.50",
	"Opera/9.80 (X11; Linux i686; U; hu) Presto/2.9.168 Version/11.50",
	"Opera/9.80 (X11; Linux i686; U; ru) Presto/2.8.131 Version/11.11",
	"Opera/9.80 (X11; Linux i686; U; es-ES) Presto/2.8.131 Version/11.11",
	"Mozilla/5.0 (Windows NT 5.1; U; en; rv:1.8.1) Gecko/20061208 Firefox/5.0 Opera 11.11",
	"Opera/9.80 (X11; Linux x86_64; U; bg) Presto/2.8.131 Version/11.10",
	"Opera/9.80 (Windows NT 6.0; U; en) Presto/2.8.99 Version/11.10",
	"Opera/9.80 (Windows NT 5.1; U; zh-tw) Presto/2.8.131 Version/11.10",
	"Opera/9.80 (Windows NT 6.1; Opera Tablet/15165; U; en) Presto/2.8.149 Version/11.1",
	"Opera/9.80 (X11; Linux x86_64; U; Ubuntu/10.10 (maverick); pl) Presto/2.7.62 Version/11.01",
	"Opera/9.80 (X11; Linux i686; U; ja) Presto/2.7.62 Version/11.01",
	"Opera/9.80 (X11; Linux i686; U; fr) Presto/2.7.62 Version/11.01",
	"Opera/9.80 (Windows NT 6.1; U; zh-tw) Presto/2.7.62 Version/11.01",
	"Opera/9.80 (Windows NT 6.1; U; zh-cn) Presto/2.7.62 Version/11.01",
	"Opera/9.80 (Windows NT 6.1; U; sv) Presto/2.7.62 Version/11.01",
	"Opera/9.80 (Windows NT 6.1; U; en-US) Presto/2.7.62 Version/11.01",
	"Opera/9.80 (Windows NT 6.1; U; cs) Presto/2.7.62 Version/11.01",
	"Opera/9.80 (Windows NT 6.0; U; pl) Presto/2.7.62 Version/11.01",
	"Opera/9.80 (Windows NT 5.2; U; ru) Presto/2.7.62 Version/11.01",
	"Opera/9.80 (Windows NT 5.1; U;) Presto/2.7.62 Version/11.01",
}
