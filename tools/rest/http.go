package rest

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/zhulingbiezhi/go12306/tools/errors"
	"github.com/zhulingbiezhi/go12306/tools/logger"
)

type RestHttp struct {
	err         error
	header      *http.Header
	cookies     []*http.Cookie
	body        string
	respBody    []byte
	respHeader  *http.Header
	respCookies []*http.Cookie
	logOp       *LogOption
	done        bool //是否已发起请求
}

func NewHttp() *RestHttp {
	return &RestHttp{
		header:  &http.Header{},
		cookies: make([]*http.Cookie, 0),
		//logOp:   &LogOption{LogBody: true, LogHeader: true},
	}
}

func (r *RestHttp) SetHeader(head map[string]interface{}) *RestHttp {
	if r.err != nil {
		return r
	}
	for key, value := range head {
		r.header.Set(key, fmt.Sprint(value))
	}
	return r
}

type RestOption func(*RestHttp)

func RestMultiCookiesOption(ck []*http.Cookie) RestOption {
	return func(r *RestHttp) {
		r.cookies = append(r.cookies, ck...)
	}
}

func RestCookieKVOption(kv map[string]interface{}) RestOption {
	return func(r *RestHttp) {
		for key, value := range kv {
			r.cookies = append(r.cookies, &http.Cookie{
				Name:  key,
				Value: fmt.Sprint(value),
			})
		}
	}
}

type LogOption struct {
	LogBody   bool
	LogHeader bool
}

func (r *RestHttp) SetRestLogOption(o *LogOption) *RestHttp {
	r.logOp = o
	return r
}

func (r *RestHttp) SetCookie(ops ...RestOption) *RestHttp {
	if r.err != nil {
		return r
	}
	for _, op := range ops {
		op(r)
	}
	return r
}

func (r *RestHttp) SetContentType(c string) *RestHttp {
	if r.err != nil {
		return r
	}
	r.header.Set("Content-Type", c)
	return r
}

func (r *RestHttp) DoRest(method, url string, body interface{}) *RestHttp {
	_, err := r.Do(method, url, body)
	if err != nil {
		r.err = errors.Errorf(err, "rest.Do err")
	}
	return r
}

func (r *RestHttp) Do(method, url string, body interface{}) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.logOp != nil {
		logger.Debug("[rest-request-url] ", url)
	}
	var reader io.Reader
	if body != nil {
		switch tt := body.(type) {
		case []byte:
			reader = bytes.NewReader(tt)
			r.body = string(tt)
		case string:
			reader = strings.NewReader(tt)
			r.body = tt
		default:
			return nil, errors.Errorf(nil, "rest.Do not support body type %T", body)
		}
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, errors.Errorf(err, "http.NewRequest err")
	}
	req.Header = *r.header
	for i := range r.cookies {
		if r.cookies[i] != nil {
			req.AddCookie(r.cookies[i])
		}
	}
	if r.logOp != nil && r.logOp.LogHeader {
		logger.Debugf("[rest-request-header] %+v", req.Header)
	}
	if r.logOp != nil && r.logOp.LogBody {
		logger.Debug("[rest-request-body] ", r.body)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Errorf(err, "http.DefaultClient.Do err")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf(nil, "http status code %s != 200", resp.StatusCode)
	}
	defer resp.Body.Close()
	r.respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Errorf(err, "ioutil.ReadAll err")
	}
	if r.logOp != nil && r.logOp.LogHeader {
		logger.Debugf("[rest-response-header] %+v", resp.Header)
	}
	if r.logOp != nil && r.logOp.LogBody {
		if !strings.Contains(resp.Header.Get("Content-Type"), "html") {
			logger.Debugf("[rest-response-body] %+v", string(r.respBody))
		} else {
			logger.Debugf("[rest-response-body] response is html, didn't show log")
		}
	}
	r.respHeader = &resp.Header
	r.respCookies = resp.Cookies()
	return r.respBody, nil
}

func (r *RestHttp) ParseJsonBody(result interface{}) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	err := json.Unmarshal(r.respBody, result)
	if err != nil {
		return nil, errors.Errorf(err, "json.Unmarshal err")
	}
	return r.respBody, nil
}

func (r *RestHttp) ParseXmlBody(result interface{}) ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	err := xml.Unmarshal(r.respBody, result)
	if err != nil {
		return nil, errors.Errorf(err, "xml.Unmarshal err")
	}
	return r.respBody, nil
}

func (r *RestHttp) AutoParseBody(result interface{}) ([]byte, error) {
	contentType := ""
	switch {
	case r.respHeader.Get("Content-Type") != "":
		contentType = r.respHeader.Get("Content-Type")
	case r.respHeader.Get("Content-Type") != "":
		contentType = r.respHeader.Get("Content-Type")
	default:
		contentType = ContentTypeJSON
	}
	switch {
	case strings.Contains(contentType, "application/json"):
		return r.ParseJsonBody(result)
	case strings.Contains(contentType, "application/xml"):
		return r.ParseJsonBody(result)
	}
	return r.respBody, errors.Errorf(nil, "not support content type %s", contentType)
}

func (r *RestHttp) RespCookies() map[string]*http.Cookie {
	cookieMap := make(map[string]*http.Cookie)
	for i, cookie := range r.respCookies {
		cookieMap[cookie.Name] = r.respCookies[i]
	}
	return cookieMap
}
