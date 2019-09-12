package goribotExts

import (
	"crypto/md5"
	"github.com/zhshch2002/goribot"
	"net/url"
	"sort"
	"strings"
	"sync"
)

type DeduplicatePipeline struct {
	goribot.Pipeline
	CrawledHash map[[md5.Size]byte]struct{}
	lock        sync.Mutex
}

func NewDeduplicatePipeline() *DeduplicatePipeline {
	return &DeduplicatePipeline{}
}

func (s *DeduplicatePipeline) Init(spider *goribot.Spider) {
	s.CrawledHash = make(map[[md5.Size]byte]struct{})
	s.lock = sync.Mutex{}
}
func (s *DeduplicatePipeline) OnRequest(spider *goribot.Spider, request *goribot.Request) *goribot.Request {
	has := GetRequestHash(request)
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.CrawledHash[has]; ok {
		return nil
	}

	s.CrawledHash[has] = struct{}{}
	return request
}
func (s *DeduplicatePipeline) OnResponse(spider *goribot.Spider, response *goribot.Response) *goribot.Response {
	return response
}
func (s *DeduplicatePipeline) OnItem(spider *goribot.Spider, item interface{}) interface{} {
	return item
}
func (s *DeduplicatePipeline) OnError(spider *goribot.Spider, err error) {}

func (s *DeduplicatePipeline) Finish(spider *goribot.Spider) {}

func GetRequestHash(r *goribot.Request) [md5.Size]byte {
	u := r.Url
	UrtStr := u.Scheme + "://"
	if u.User != nil {
		UrtStr += u.User.String() + "@"
	}
	UrtStr += strings.ToLower(u.Host)
	path := u.EscapedPath()
	if path != "" && path[0] != '/' {
		UrtStr += "/"
	}
	UrtStr += path
	if u.RawQuery != "" {
		QueryParam := u.Query()
		var QueryK []string
		for k, _ := range QueryParam {
			QueryK = append(QueryK, k)
		}
		sort.Strings(QueryK)
		var QueryStrList []string
		for _, k := range QueryK {
			val := QueryParam[k]
			sort.Strings(val)
			for _, v := range val {
				QueryStrList = append(QueryStrList, url.QueryEscape(k)+"="+url.QueryEscape(v))
			}
		}
		UrtStr += "?" + strings.Join(QueryStrList, "&")
	}

	Header := r.Header
	var HeaderK []string
	for k, _ := range Header {
		HeaderK = append(HeaderK, k)
	}
	sort.Strings(HeaderK)
	var HeaderStrList []string
	for _, k := range HeaderK {
		val := Header[k]
		sort.Strings(val)
		for _, v := range val {
			HeaderStrList = append(HeaderStrList, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}
	HeaderStr := strings.Join(HeaderStrList, "&")

	data := []byte(UrtStr + HeaderStr)
	data = append(data, r.Body...)
	has := md5.Sum(data)
	return has
}
