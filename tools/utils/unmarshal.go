package utils

import (
	"encoding/json"
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/zhulingbiezhi/go12306/pkg/helper"
	"github.com/zhulingbiezhi/go12306/tools/logger"

	"io/ioutil"
	"net/http"
)

//ParseBodyAndUnmarshal 得到POST过来的body并json解码的对象
func ParseBodyAndUnmarshal(r *http.Request, req interface{}) error {
	if r == nil {
		return fmt.Errorf(" http.Request为空.URL:%s", r.RequestURI)
	}

	if r.Body == nil {
		return fmt.Errorf(" r.Body为空.URL:%s", r.RequestURI)
	}

	postData, er := ioutil.ReadAll(r.Body)
	if er != nil {
		return fmt.Errorf(" ReadAll发生异常. URL:%s err:%s", r.RequestURI, er)
	}
	defer r.Body.Close()

	if len(postData) == 0 {
		return fmt.Errorf(" postData为空.  URL:%s r.Body为空.", r.RequestURI)
	}
	logger.Info("request_body:", string(postData))
	if er = json.Unmarshal(postData, req); er != nil {
		return fmt.Errorf(" json解码失败. err:%s, body:%s", er, string(postData))
	}
	if parse, ok := req.(helper.ParseRequest); ok {
		if err := parse.ParseRequest(r); err != nil {
			return fmt.Errorf("ParseRequest err: %s", err.Error())
		}
	}
	if valid, ok := req.(helper.ValidRequest); ok {
		if er := valid.ValidParam(); er != nil {
			return fmt.Errorf(" 参数错误. err:%s, body:%s", er, string(postData))
		}
	} else {
		if flag, er := govalidator.ValidateStruct(r); er != nil {
			return fmt.Errorf(" 参数错误. err:%s, body:%s", er, string(postData))
		} else if !flag {
			return fmt.Errorf(" 参数错误. err:%s, body:%s", fmt.Errorf("验证未通过"), string(postData))
		}
	}
	return nil
}
