package code

import (
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/zhulingbiezhi/go12306/pkg/helper"
	"github.com/zhulingbiezhi/go12306/tools/errors"
	"github.com/zhulingbiezhi/go12306/tools/logger"
	"github.com/zhulingbiezhi/go12306/tools/rest"
	"github.com/zhulingbiezhi/go12306/tools/utils"
)

type Code struct {
	CookieMap map[string]*http.Cookie
}

type Request struct {
	Vals url.Values
}

type imageResponse struct {
	Image         string `json:"image"`
	ResultMessage string `json:"result_message"`
	ResultCode    string `json:"result_code"`
}

func GetAuthCode(ctx context.Context, vals url.Values, cookieMap map[string]*http.Cookie) (string, error) {
	if len(vals) == 0 {
		return "", errors.Errorf(nil, "data is empty")
	}
	c := Code{
		CookieMap: cookieMap,
	}
	resp, err := c.generatorCodeImage(ctx, vals)
	if err != nil {
		return "", errors.Errorf(err, "generatorCodeImage err")
	}
	//save picture
	if false {
		if err := saveImage(resp.Image); err != nil {
			return "", errors.Errorf(err, "saveImage err")
		}
	}
	//image recognition
	ret, err := getVerifyCode(resp.Image)
	if err != nil {
		return "", errors.Errorf(err, "GetVerifyCode err")
	}
	//check auth code
	answer := strings.Join(getImagePositionByOffset(ret), ",")
	success, err := c.checkAuthCode(ctx, answer)
	if err != nil {
		return "", errors.Errorf(err, "CheckAuthCode err")
	}
	if !success {
		return "", errors.Errorf(nil, "CheckAuthCode fail")
	}
	return answer, nil
}

func (c *Code) generatorCodeImage(ctx context.Context, vals url.Values) (*imageResponse, error) {
	urlStr := helper.API_AUTH_CODE_BASE64_DOWNLOAD + "?" + vals.Encode()
	code := &imageResponse{}
	rs := rest.NewHttp()
	_, err := rs.DoRest(http.MethodGet, urlStr, nil).ParseJsonBody(code)
	if err != nil {
		return nil, err
	}

	for key, ck := range rs.RespCookies() {
		switch key {
		case helper.Cookie_PassportCt, helper.Cookie_PassportSession:
			c.CookieMap[ck.Name] = rs.RespCookies()[key]
		}
	}

	if code.ResultCode != "0" || code.ResultMessage != "生成验证码成功" {
		return nil, errors.Errorf(nil, "code: %s, message: %s", code.ResultCode, code.ResultMessage)
	}
	return code, nil
}

func getVerifyCode(img string) ([]int, error) {
	vals := make(url.Values, 0)
	vals.Set("img", img)
	ret := struct {
		Msg    string `json:"msg"`
		Result []int  `json:"result"`
	}{}
	rs := rest.NewHttp().SetContentType(rest.ContentTypeForm)
	_, err := rs.DoRest(http.MethodPost, helper.API_FREE_CODE_QCR_API_URL, vals.Encode()).ParseJsonBody(&ret)
	if err != nil {
		return nil, err
	}
	if ret.Msg != "success" || len(ret.Result) == 0 {
		return nil, errors.Errorf(nil, "result is not success %+v", ret)
	}
	return ret.Result, nil
}

func (c *Code) checkAuthCode(ctx context.Context, answer string) (bool, error) {
	vals := url.Values{
		"login_site": []string{"E"},
		"module":     []string{"login"},
		"rand":       []string{"sjrand"},
		"_":          []string{strconv.FormatFloat(rand.Float64(), 'f', -1, 64)},
		"answer":     []string{answer},
	}
	urlStr := helper.API_AUTH_CODE_CHECK_URL + "?" + vals.Encode()
	ret := struct {
		ResultMessage string `json:"result_message"`
		ResultCode    string `json:"result_code"`
	}{}
	rs := rest.NewHttp().SetContentType(rest.ContentTypeForm)
	rs.SetCookie(rest.RestMultiCookiesOption([]*http.Cookie{
		c.CookieMap[helper.Cookie_PassportSession],
		c.CookieMap[helper.Cookie_PassportCt],
	}))
	_, err := rs.DoRest(http.MethodGet, urlStr, nil).ParseJsonBody(&ret)
	if err != nil {
		return false, err
	}
	bSuccess := ret.ResultCode == "4"
	if bSuccess {
		logger.Info("check auth code success")
	}
	return bSuccess, nil
}

func getImagePositionByOffset(offsets []int) []string {
	positions := make([]string, 0)
	width := (75)
	height := (75)
	for _, offset := range offsets {
		random_x := rand.Intn(20) + -10
		random_y := rand.Intn(20) + -10
		offset = int(offset)
		x := width*((offset-1)%4+1) - width/2 + random_x
		y := height*int(math.Ceil(float64(offset)/4)) - height/2 + random_y

		positions = append(positions, strconv.Itoa(x))
		positions = append(positions, strconv.Itoa(y))
	}
	return positions
}

func saveImage(imgStr string) error {
	//save picture
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgStr))
	img, _, err := image.Decode(reader)
	if err != nil {
		return errors.Errorf(err, "image.Decode err")
	}
	dir := utils.GetCurDir()
	//Encode from image format to writer
	dstFileName := dir + fmt.Sprintf("/code_%d.jpeg", 0)
	f, err := os.OpenFile(dstFileName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	err = jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
	if err != nil {
		return err
	}
	return nil
}
