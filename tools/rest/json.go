package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func MarshalAPIJSON(code, message string, success bool,
	xStruct interface{}) ([]byte, error) {
	x := &APIJSONFormat{}
	x.GError.Code = code
	x.GError.Message = message
	x.Success = success
	x.XStruct = xStruct

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)

	if err := enc.Encode(x); err != nil {
		return nil, fmt.Errorf("[MarshalJSON] Marshal失败! err:%s", err.Error())
	}
	return buf.Bytes(), nil
}

type APIJSONFormat struct {
	GError  `json:"error,omitempty"`
	XStruct interface{} `json:"result"  structs:"result"`
	Success bool        `json:"success"  structs:"success"`
}

//GError 错误信息，succes为true时，值都为空
type GError struct {
	Code    string `json:"code"  structs:"code"`
	Message string `json:"message"  structs:"message"`
}

//GetSuccessJSON 执行成功，返回的处理
func GetSuccessJSON(xStruct interface{}) ([]byte, error) {
	return MarshalAPIJSON("", "", true, xStruct)
}

//GetFailedJSON 将错误信息依指定格式序列化
func GetFailedJSON(code, message string) ([]byte, error) {
	return MarshalAPIJSON(code, message, false, struct{}{})
}
