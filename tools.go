package goribot

import (
	"bytes"
	"crypto/md5"
	"golang.org/x/net/html/charset"
	"io/ioutil"
	"net/url"
	"sort"
	"strings"
)

func encodeBytes(b []byte, contentType string) ([]byte, error) {
	r, err := charset.NewReader(bytes.NewReader(b), contentType)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(r)
}

// GetRequestHash return a hash of url,header,cookie and body data from a request
func GetRequestHash(r *Request) [md5.Size]byte {
	u := r.URL
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
	for _, i := range r.Cookies() {
		Cookie = append(Cookie, i.Name+"="+i.Value)
	}
	CookieStr := strings.Join(Cookie, "&")

	data := []byte(strings.Join([]string{UrtStr, HeaderStr, CookieStr}, "@#@"))
	data = append(data, r.GetBody()...)
	has := md5.Sum(data)
	return has
}
