package cookie

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/zhulingbiezhi/go12306/pkg/helper"
	"github.com/zhulingbiezhi/go12306/tools/errors"
	"github.com/zhulingbiezhi/go12306/tools/logger"
	"github.com/zhulingbiezhi/go12306/tools/rest"
)

func GetRailDevice(ctx context.Context) error {
	retryTimes := 0
retry:
	retryTimes++
	if retryTimes > 5 {
		return errors.Errorf(nil, "重试超过5次")
	}
	result, err := getRailDevice(ctx)
	if err != nil {
		logger.Error("getRailDevice err", err)
		goto retry
	}
	index := strings.Index(result, "callbackFunction")
	if index > 0 {
		logger.Error("not found callbackFunction")
		time.Sleep(time.Millisecond * 200)
		goto retry
	}
	ret := struct {
		Exp string `json:"exp"`
		Dfp string `json:"dfp"`
	}{}
	result = result[18 : len(result)-2]
	err = json.Unmarshal([]byte(result), &ret)
	if err != nil {
		return errors.Errorf(err, "json.Unmarshal err")
	}
	cookie, ok := ctx.Value("cookie").(map[string]*http.Cookie)
	if ok {
		cookie[helper.Cookie_RAIL_DEVICEID] = &http.Cookie{
			Name:  helper.Cookie_RAIL_DEVICEID,
			Value: ret.Dfp,
		}
		cookie[helper.Cookie_RAIL_EXPIRATION] = &http.Cookie{
			Name:  helper.Cookie_RAIL_EXPIRATION,
			Value: ret.Exp,
		}
	}
	return nil
}

func getRailDevice(ctx context.Context) (string, error) {
	res := struct {
		ID string `json:"id"`
	}{}
	_, err := rest.NewHttp().DoRest(http.MethodGet, helper.API_GET_BROWSER_DEVICE_ID_URL, nil).ParseJsonBody(&res)
	if err != nil {
		return "", err
	}
	urlBytes, err := base64.StdEncoding.DecodeString(res.ID)
	if err != nil {
		return "", errors.Errorf(err, "base64.StdEncoding.DecodeString err")
	}
	b, err := rest.NewHttp().Do(http.MethodGet, string(urlBytes), nil)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
