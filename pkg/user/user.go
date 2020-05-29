package user

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/mitchellh/mapstructure"
	"github.com/zhulingbiezhi/go12306/pkg/helper"
	"github.com/zhulingbiezhi/go12306/pkg/helper/code"
	"github.com/zhulingbiezhi/go12306/tools/conf"
	"github.com/zhulingbiezhi/go12306/tools/errors"
	"github.com/zhulingbiezhi/go12306/tools/logger"
	"github.com/zhulingbiezhi/go12306/tools/rest"
)

const (
	maxRetryTime = 5
)

type User struct {
	Name         string
	Uuid         string    `mapstructure:"uuid"`
	UserName     string    `mapstructure:"user_name"`
	UserPassword string    `mapstructure:"user_password"`
	Members      []*Member `json:"members"`
}

type Member struct {
	Name string `json:"name"`
	Type int    `json:"type"`
}

var userMap = make(map[string]*User)

func init() {
	for _, user := range conf.Conf.Users {
		u := &User{}
		err := mapstructure.Decode(user, u)
		if err != nil {
			logger.Error("mapstructure.Decode err", err)
		} else {
			if _, ok := userMap[u.Uuid]; ok {
				logger.Errorf("user %s already exist", u.Uuid)
			} else {
				userMap[u.Uuid] = u
			}
		}
	}
	logger.Info("users---", awsutil.Prettify(userMap))
}

func GetUser(uuid string) (*User, error) {
	user, ok := userMap[uuid]
	if !ok {
		return nil, fmt.Errorf("user uuid: %s is empty", uuid)
	}
	return user, nil
}

func (u *User) Login(ctx context.Context) error {
	uamtk, err := u.login(ctx)
	if err != nil {
		return errors.Errorf(err, "user login err")
	}
	tk, err := UAMtk(ctx, uamtk)
	if err != nil {
		return errors.Errorf(err, "UAMtk err")
	}
	err = UAMAuthClient(ctx, tk)
	if err != nil {
		return errors.Errorf(err, "UAMAuthClient err")
	}
	return nil
}

func (u *User) login(ctx context.Context) (string, error) {
	logger.Infof("login user: %s", u.Uuid)
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
	answer, err := code.GetAuthCode(ctx, vals)
	if err != nil {
		logger.Errorf("GetAuthCode %d err: %+v", retryTimes, err)
		retryTimes++
		goto retry
	}
	ret := struct {
		ResultCode    int    `json:"result_code"`
		ResultMessage string `json:"result_message"`
		UaMTK         string `json:"uamtk"`
	}{}
	v := url.Values{
		"username": []string{u.UserName},
		"password": []string{u.UserPassword},
		"appid":    []string{"otn"},
		"answer":   []string{answer},
	}
	rs := rest.NewHttp().SetContentType(rest.ContentTypeForm)
	cookie, ok := ctx.Value("cookie").(map[string]*http.Cookie)
	if ok {
		rs.SetCookie(rest.RestMultiCookiesOption([]*http.Cookie{
			cookie[helper.Cookie_PassportCt],
			cookie[helper.Cookie_PassportSession],
		}))
	}
	rs.SetHeader(map[string]interface{}{
		helper.Header_USER_AGENT: helper.UserAgentChrome,
	})
	rs.SetCookie(rest.RestCookieKVOption(map[string]interface{}{
		helper.Cookie_RAIL_EXPIRATION: conf.Conf.RailExpire,
		helper.Cookie_RAIL_DEVICEID:   conf.Conf.RailDevice,
	}))
	b, err := rs.DoRest(http.MethodPost, conf.API_BASE_LOGIN_URL, v.Encode()).ParseJsonBody(&ret)
	if err != nil {
		return "", errors.Errorf(err, "")
	}
	if ret.ResultCode != 0 {
		return "", errors.Errorf(nil, "login fail: %+s", string(b))
	}
	logger.Infof("login user: %s success", u.Uuid)
	return ret.UaMTK, nil
}

type UAMtkResponse struct {
	Apptk         interface{} `json:"apptk"`
	ResultMessage string      `json:"result_message"`
	ResultCode    int         `json:"result_code"`
	Newapptk      string      `json:"newapptk"`
}

func UAMtk(ctx context.Context, uamtk string) (string, error) {
	rs := rest.NewHttp().SetContentType(rest.ContentTypeForm)
	rs.SetHeader(map[string]interface{}{
		"Referer":                "https://kyfw.12306.cn/otn/passport?redirect=/otn/login/userLogin",
		"Origin":                 "https://kyfw.12306.cn",
		helper.Header_USER_AGENT: helper.UserAgentChrome,
	})
	rs.SetCookie(rest.RestCookieKVOption(map[string]interface{}{
		helper.Cookie_Uamtk:           uamtk,
		helper.Cookie_RAIL_EXPIRATION: conf.Conf.RailExpire,
		helper.Cookie_RAIL_DEVICEID:   conf.Conf.RailDevice,
	}))
	ck, ok := ctx.Value("cookie").(map[string]*http.Cookie)
	if ok {
		rs.SetCookie(rest.RestMultiCookiesOption([]*http.Cookie{
			ck[helper.Cookie_PassportCt],
			ck[helper.Cookie_PassportSession],
		}))
	}
	ret := UAMtkResponse{}
	vals := url.Values{
		"appid": []string{"otn"},
	}
	_, err := rs.DoRest(http.MethodPost, conf.API_AUTH_UAMTK_URL, vals.Encode()).ParseJsonBody(&ret)
	if err != nil {
		return "", err
	}
	ck[helper.Cookie_Uamtk] = rs.RespCookies()[helper.Cookie_Uamtk]
	return ret.Newapptk, nil
}

type UAMAuthClientResponse struct {
	ResultCode    int    `json:"result_code"`
	ResultMessage string `json:"result_message"`
	Username      string `json:"username"`
	Apptk         string `json:"apptk"`
}

func UAMAuthClient(ctx context.Context, tk string) error {
	ret := UAMAuthClientResponse{}
	rs := rest.NewHttp().SetContentType(rest.ContentTypeForm)
	rs.SetHeader(map[string]interface{}{
		"Referer":                "https://kyfw.12306.cn/otn/passport?redirect=/otn/login/userLogin",
		"Origin":                 "https://kyfw.12306.cn",
		helper.Header_USER_AGENT: helper.UserAgentChrome,
	})
	vals := make(url.Values)
	vals.Set("tk", tk)

	rs.SetCookie(rest.RestCookieKVOption(map[string]interface{}{
		helper.Cookie_RAIL_EXPIRATION: conf.Conf.RailExpire,
		helper.Cookie_RAIL_DEVICEID:   conf.Conf.RailDevice,
	}))
	_, err := rs.DoRest(http.MethodPost, conf.API_AUTH_UAMAUTHCLIENT_URL, vals.Encode()).ParseJsonBody(&ret)
	if err != nil {
		return err
	}
	ck, ok := ctx.Value("cookie").(map[string]*http.Cookie)
	if ok {
		ck[helper.Cookie_Apptk] = rs.RespCookies()[helper.Cookie_Apptk]
	}
	return nil
}
