package account

import (
	"context"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"

	"github.com/zhulingbiezhi/go12306/config"
	"github.com/zhulingbiezhi/go12306/pkg/captcha"
	"github.com/zhulingbiezhi/go12306/pkg/common"
	"github.com/zhulingbiezhi/go12306/tools/errors"
	"github.com/zhulingbiezhi/go12306/tools/logger"
	"github.com/zhulingbiezhi/go12306/tools/rest"
)

const (
	maxRetryTime = 5
)

type Account struct {
	Name            string
	AccountName     string
	AccountPassword string
	members         []*Member
	cookieMap       map[string]*http.Cookie
}

type Member struct {
	Name string `json:"name"`
	Type int    `json:"type"`
}

func (u *Account) Login(ctx context.Context) error {
	uamtk, err := u.login(ctx)
	if err != nil {
		return errors.Errorf(err, "account login err")
	}
	tk, err := u.uamtk(ctx, uamtk)
	if err != nil {
		return errors.Errorf(err, "UAMtk err")
	}
	err = u.uamAuthClient(ctx, tk)
	if err != nil {
		return errors.Errorf(err, "UAMAuthClient err")
	}
	return nil
}

type loginResponse struct {
	ResultCode    int    `json:"result_code"`
	ResultMessage string `json:"result_message"`
	UaMTK         string `json:"uamtk"`
}

func (u *Account) login(ctx context.Context) (string, error) {
	retryTimes := 0
retry:
	if retryTimes > maxRetryTime {
		return "", errors.Errorf(nil, "login fail, GetAuthCode reach max retry times")
	}
	vals := url.Values{
		"login_site": []string{"E"},
		"module":     []string{"login"},
		"rand":       []string{"sjrand"},
		"_":          []string{strconv.FormatFloat(rand.Float64(), 'f', -1, 64)},
	}
	answer, err := captcha.GetAuthCode(ctx, vals, u.cookieMap)
	if err != nil {
		logger.Errorf("GetAuthCode %d err: %+v", retryTimes, err)
		retryTimes++
		goto retry
	}

	v := url.Values{
		"username": []string{u.AccountName},
		"password": []string{u.AccountPassword},
		"appid":    []string{"otn"},
		"answer":   []string{answer},
	}
	rs := rest.NewHttp().SetContentType(rest.ContentTypeForm)
	rs.SetCookie(rest.RestMultiCookiesOption([]*http.Cookie{
		u.cookieMap[common.Cookie_PassportCt],
		u.cookieMap[common.Cookie_PassportSession],
	}))
	rs.SetCookie(rest.RestCookieKVOption(map[string]interface{}{
		common.Cookie_RAIL_EXPIRATION: config.Conf.RailExpire,
		common.Cookie_RAIL_DEVICEID:   config.Conf.RailDevice,
	}))
	rs.SetHeader(map[string]interface{}{
		common.Header_USER_AGENT: common.UserAgentChrome,
	})
	ret := loginResponse{}
	b, err := rs.DoRest(http.MethodPost, common.API_BASE_LOGIN_URL, v.Encode()).ParseJsonBody(&ret)
	if err != nil {
		return "", errors.Errorf(err, "")
	}
	if ret.ResultCode != 0 {
		return "", errors.Errorf(nil, "login fail: %+s", string(b))
	}
	return ret.UaMTK, nil
}

type uamtkResponse struct {
	Apptk         interface{} `json:"apptk"`
	ResultMessage string      `json:"result_message"`
	ResultCode    int         `json:"result_code"`
	Newapptk      string      `json:"newapptk"`
}

func (u *Account) uamtk(ctx context.Context, uamtk string) (string, error) {
	rs := rest.NewHttp().SetContentType(rest.ContentTypeForm)
	rs.SetHeader(map[string]interface{}{
		"Referer":                common.BASE_URL_OF_12306 + "/otn/passport?redirect=/otn/login/userLogin",
		"Origin":                 common.BASE_URL_OF_12306,
		common.Header_USER_AGENT: common.UserAgentChrome,
	})
	rs.SetCookie(rest.RestCookieKVOption(map[string]interface{}{
		common.Cookie_Uamtk:           uamtk,
		common.Cookie_RAIL_EXPIRATION: config.Conf.RailExpire,
		common.Cookie_RAIL_DEVICEID:   config.Conf.RailDevice,
	}))
	rs.SetCookie(rest.RestMultiCookiesOption([]*http.Cookie{
		u.cookieMap[common.Cookie_PassportCt],
		u.cookieMap[common.Cookie_PassportSession],
	}))

	ret := uamtkResponse{}
	vals := url.Values{
		"appid": []string{"otn"},
	}
	_, err := rs.DoRest(http.MethodPost, common.API_AUTH_UAMTK_URL, vals.Encode()).ParseJsonBody(&ret)
	if err != nil {
		return "", err
	}
	u.cookieMap[common.Cookie_Uamtk] = rs.RespCookies()[common.Cookie_Uamtk]
	return ret.Newapptk, nil
}

type uamAuthClientResponse struct {
	ResultCode    int    `json:"result_code"`
	ResultMessage string `json:"result_message"`
	Username      string `json:"username"`
	Apptk         string `json:"apptk"`
}

func (u *Account) uamAuthClient(ctx context.Context, tk string) error {
	ret := uamAuthClientResponse{}
	rs := rest.NewHttp().SetContentType(rest.ContentTypeForm)
	rs.SetHeader(map[string]interface{}{
		"Referer":                common.BASE_URL_OF_12306 + "/otn/passport?redirect=/otn/login/userLogin",
		"Origin":                 common.BASE_URL_OF_12306,
		common.Header_USER_AGENT: common.UserAgentChrome,
	})
	vals := make(url.Values)
	vals.Set("tk", tk)
	rs.SetCookie(rest.RestCookieKVOption(map[string]interface{}{
		common.Cookie_RAIL_EXPIRATION: config.Conf.RailExpire,
		common.Cookie_RAIL_DEVICEID:   config.Conf.RailDevice,
	}))
	_, err := rs.DoRest(http.MethodPost, common.API_AUTH_UAMAUTHCLIENT_URL, vals.Encode()).ParseJsonBody(&ret)
	if err != nil {
		return err
	}
	u.cookieMap[common.Cookie_Apptk] = rs.RespCookies()[common.Cookie_Apptk]
	return nil
}
