package rest

import (
	"fmt"
	"github.com/zhulingbiezhi/go12306/helpers/logger"
	"net/http"
)

const (
	//ContentTypeJSON json类型
	ContentTypeJSON = "application/json; charset=UTF-8"
	//ContentTypeXML xml类型
	ContentTypeXML = "application/xml; charset=UTF-8"
	//ContentTypeForm form表单 application/x-www-form-urlencoded;charset=utf-8
	ContentTypeForm = "application/x-www-form-urlencoded; charset=UTF-8"
)

//ToFailedJSON 返回处理失败的JSON给调用方
func ToFailedJSON(w http.ResponseWriter, code, message string) {
	body, err := GetFailedJSON(code, message)
	if err != nil {
		logger.Error(" [ToFailedJSON] err:", err,
			" code:", code, " message:", message)
		return
	}
	logger.Error("[ToFailedJSON]", string(body))
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		logger.Error("[ToFailedJSON] error", err)
		return
	}
}

func ToUnauthJSON(w http.ResponseWriter, code, message string) {
	body, err := GetFailedJSON(code, message)
	if err != nil {
		logger.Error(" [ToUnauthJSON] err:", err,
			" code:", code, " message:", message)
		return
	}
	logger.Info("[ToUnauthJSON]", string(body))
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(http.StatusUnauthorized)
	_, err = w.Write(body)
	if err != nil {
		logger.Error("[ToFailedJSON] error", err)
		return
	}
}

func ToSuccessJSON(w http.ResponseWriter, xStruct interface{}) {
	body, err := GetSuccessJSON(xStruct)
	if err != nil {
		logger.Error("[ToSuccessJSON] GetSuccessJSON error", err, fmt.Sprintf("struct: %+v", xStruct))
		return
	}
	logger.Info("[ToSuccessJSON]", string(body))
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		logger.Error("[ToSuccessJSON] Write body error:", err, "body: ", string(body))
		return
	}
	return
}
