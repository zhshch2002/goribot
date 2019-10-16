package goribot

import (
	"crypto/md5"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// NewGetReq create a new get request
func NewGetReq(rawurl string) (*Request, error) {
	req := NewRequest()
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	req.Url = u
	req.Method = http.MethodGet
	return req, nil
}

// MustNewGetReq  create a new get request,if get error will do panic
func MustNewGetReq(rawurl string) *Request {
	res, err := NewGetReq(rawurl)
	if err != nil {
		panic(err)
	}
	return res
}

// NewPostReq create a new post request
func NewPostReq(rawurl string, datatype PostDataType, rawdata interface{}) (*Request, error) {
	req := NewRequest()
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	req.Url = u
	req.Method = http.MethodPost

	var data []byte
	ct := ""
	switch datatype {
	case TextPostData:
		ct = "text/plain"
		data = []byte(rawdata.(string))
		break
	case UrlencodedPostData:
		ct = "application/x-www-form-urlencoded"
		var urlS url.URL
		q := urlS.Query()
		for k, v := range rawdata.(map[string]string) {
			q.Add(k, v)
		}
		data = []byte(q.Encode())
		break
	case JsonPostData:
		ct = "application/json"
		tdata, err := json.Marshal(rawdata)
		if err != nil {
			return nil, err
		}
		data = tdata
		break
	}

	req.SetHeader("Content-Type", ct).SetBody(data)

	return req, nil
}

// MustNewPostReq create a new post request,if get error will do panic
func MustNewPostReq(rawurl string, datatype PostDataType, rawdata interface{}) *Request {
	res, err := NewPostReq(rawurl, datatype, rawdata)
	if err != nil {
		panic(err)
	}
	return res
}

// GetRequestHash return a hash of url,header,cookie and body data from a request
func GetRequestHash(r *Request) [md5.Size]byte {
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
		for k := range QueryParam {
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
	for k := range Header {
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

	Cookie := []string{}
	for _, i := range r.Cookie {
		Cookie = append(Cookie, i.Name+"="+i.Value)
	}
	CookieStr := strings.Join(Cookie, "&")

	data := []byte(strings.Join([]string{UrtStr, HeaderStr, CookieStr}, "@#@"))
	data = append(data, r.Body...)
	has := md5.Sum(data)
	return has
}
